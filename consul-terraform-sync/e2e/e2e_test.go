// +build e2e

package e2e

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/consul-terraform-sync/config"
	"github.com/hashicorp/consul-terraform-sync/testutils"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/require"
)

// tempDirPrefix is the prefix for the directory for a given e2e test
// where files generated from e2e are stored. This directory is
// destroyed after e2e testing if no errors.
const tempDirPrefix = "tmp_"

// resourcesDir is the sub-directory of tempDir where the
// Terraform resources created from running consul-terraform-sync are stored
const resourcesDir = "resources"

// configFile is the name of the sync config file
const configFile = "config.hcl"

func TestE2EBasic(t *testing.T) {
	// Note: no t.Parallel() for this particular test. Choosing this test to run 'first'
	// since e2e test running simultaneously will download Terraform into shared
	// directory causes some flakiness. All other e2e tests, should have t.Parallel()

	srv := newTestConsulServer(t)
	defer srv.Stop()

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "basic")
	delete := testutils.MakeTempDir(t, tempDir)
	// no defer to delete directory: only delete at end of test if no errors

	configPath := filepath.Join(tempDir, configFile)
	config := baseConfig().appendConsulBlock(srv).appendTerraformBlock(tempDir).
		appendDBTask().appendWebTask()
	config.write(t, configPath)

	err := runSyncStop(configPath, 20*time.Second)
	require.NoError(t, err)

	files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", tempDir, resourcesDir))
	require.NoError(t, err)
	require.Equal(t, 3, len(files))

	contents, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/consul_service_api.txt", tempDir, resourcesDir))
	require.NoError(t, err)
	require.Equal(t, "1.2.3.4", string(contents))

	contents, err = ioutil.ReadFile(fmt.Sprintf("%s/%s/consul_service_web.txt", tempDir, resourcesDir))
	require.NoError(t, err)
	require.Equal(t, "5.6.7.8", string(contents))

	contents, err = ioutil.ReadFile(fmt.Sprintf("%s/%s/consul_service_db.txt", tempDir, resourcesDir))
	require.NoError(t, err)
	require.Equal(t, "10.10.10.10", string(contents))

	// check statefiles exist
	status := checkStateFile(t, srv.HTTPAddr, dbTaskName)
	require.Equal(t, http.StatusOK, status)

	status = checkStateFile(t, srv.HTTPAddr, webTaskName)
	require.Equal(t, http.StatusOK, status)

	delete()
}

func TestE2ERestartSync(t *testing.T) {
	t.Parallel()

	srv := newTestConsulServer(t)
	defer srv.Stop()

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "restart")
	delete := testutils.MakeTempDir(t, tempDir)
	// no defer to delete directory: only delete at end of test if no errors

	configPath := filepath.Join(tempDir, configFile)
	config := baseConfig().appendConsulBlock(srv).appendTerraformBlock(tempDir).appendDBTask()
	config.write(t, configPath)

	err := runSyncStop(configPath, 8*time.Second)
	require.NoError(t, err)

	// rerun sync. confirm no errors e.g. recreating workspaces
	err = runSyncStop(configPath, 8*time.Second)
	require.NoError(t, err)

	delete()
}

func TestE2ERestartConsul(t *testing.T) {
	// Test that when Consul restarts, CTS is able to reconnect and react to
	// changes in Consul catalog.
	t.Parallel()

	consul := testutils.NewTestConsulServer(t, testutils.TestConsulServerConfig{
		HTTPSRelPath: "../testutils",
	})

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "restart_consul")
	cleanup := testutils.MakeTempDir(t, tempDir) // cleanup at end if no errors

	configPath := filepath.Join(tempDir, configFile)
	config := baseConfig().appendConsulBlock(consul).
		appendTerraformBlock(tempDir).appendDBTask()
	config.write(t, configPath)

	// start CTS
	cmd, err := runSync(configPath)
	require.NoError(t, err)
	defer stopCommand(cmd)
	time.Sleep(5 * time.Second)

	// stop Consul
	consul.Stop()
	time.Sleep(2 * time.Second)

	// restart Consul
	consul = testutils.NewTestConsulServer(t, testutils.TestConsulServerConfig{
		HTTPSRelPath: "../testutils",
		PortHTTPS:    consul.Config.Ports.HTTPS,
	})
	defer consul.Stop()
	time.Sleep(5 * time.Second)

	// register a new service
	apiInstance := testutil.TestService{ID: "api_new", Name: "api"}
	testutils.RegisterConsulService(t, consul, apiInstance, testutil.HealthPassing)
	time.Sleep(8 * time.Second)

	// confirm that CTS reconnected with Consul and created resource for latest service
	_, err = ioutil.ReadFile(fmt.Sprintf("%s/%s/consul_service_api_new.txt", tempDir, resourcesDir))
	require.NoError(t, err)

	cleanup()
}

func TestE2EPanosHandlerError(t *testing.T) {
	t.Parallel()

	srv := newTestConsulServer(t)
	defer srv.Stop()

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "panos_handler")
	delete := testutils.MakeTempDir(t, tempDir)
	// no defer to delete directory: only delete at end of test if no errors

	requiredProviders := `required_providers {
  panos = {
    source = "paloaltonetworks/panos"
  }
}
`
	configPath := filepath.Join(tempDir, configFile)
	config := panosBadCredConfig().appendConsulBlock(srv).
		appendTerraformBlock(tempDir, requiredProviders)
	config.write(t, configPath)

	err := runSyncOnce(configPath)
	require.Error(t, err)

	delete()
}

func TestE2ELocalBackend(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		tempDirPrefix    string
		backendConfig    string
		dbStateFilePath  string
		webStateFilePath string
	}{
		{
			"no parameters configured",
			"local_backend_default",
			`backend "local" {}`,
			fmt.Sprintf("tmp_local_backend_default/%s/terraform.tfstate.d/%s",
				dbTaskName, dbTaskName),
			fmt.Sprintf("tmp_local_backend_default/%s/terraform.tfstate.d/%s",
				webTaskName, webTaskName),
		},
		{
			"workspace_dir configured",
			"local_backend_ws_dir",
			`backend "local" {
				workspace_dir = "test-workspace"
			}`,
			fmt.Sprintf("tmp_local_backend_ws_dir/%s/test-workspace/%s",
				dbTaskName, dbTaskName),
			fmt.Sprintf("tmp_local_backend_ws_dir/%s/test-workspace/%s",
				webTaskName, webTaskName),
		},
		{
			"workspace_dir configured with tasks sharing a workspace dir",
			"local_backend_shared_ws_dir",
			`backend "local" {
				workspace_dir = "../shared-workspace"
			}`,
			fmt.Sprintf("tmp_local_backend_shared_ws_dir/shared-workspace/%s",
				dbTaskName),
			fmt.Sprintf("tmp_local_backend_shared_ws_dir/shared-workspace/%s",
				webTaskName),
		},
		{
			"path configured",
			"local_backend_path",
			`backend "local" {
				# Setting path is meaningless in Sync. TF only uses it for
				# default workspace; Sync only uses non-default workspaces. This
				# value is overridden by the workspace directory for non-default
				# workspaces.
				path = "this-will-be-replaced-by-default-dir/terraform.tfstate"
			}`,
			fmt.Sprintf("tmp_local_backend_path/%s/terraform.tfstate.d/%s",
				dbTaskName, dbTaskName),
			fmt.Sprintf("tmp_local_backend_path/%s/terraform.tfstate.d/%s",
				webTaskName, webTaskName),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestConsulServer(t)
			defer srv.Stop()

			tempDir := fmt.Sprintf("%s%s", tempDirPrefix, tc.tempDirPrefix)
			delete := testutils.MakeTempDir(t, tempDir)
			// no defer to delete directory: only delete at end of test if no errors

			config := baseConfig().appendConsulBlock(srv).
				appendTerraformBlock(tempDir, tc.backendConfig).
				appendDBTask().appendWebTask()

			configPath := filepath.Join(tempDir, configFile)
			config.write(t, configPath)

			err := runSyncOnce(configPath)
			require.NoError(t, err)

			// check that statefile was created locally
			exists, err := checkStateFileLocally(tc.dbStateFilePath)
			require.NoError(t, err)
			require.True(t, exists)

			exists, err = checkStateFileLocally(tc.webStateFilePath)
			require.NoError(t, err)
			require.True(t, exists)

			delete()
		})
	}
}

func newTestConsulServer(t *testing.T) *testutil.TestServer {
	srv := testutils.NewTestConsulServer(t, testutils.TestConsulServerConfig{
		HTTPSRelPath: "../testutils",
	})

	// Register services
	srv.AddAddressableService(t, "api", testutil.HealthPassing,
		"1.2.3.4", 8080, []string{})
	srv.AddAddressableService(t, "web", testutil.HealthPassing,
		"5.6.7.8", 8000, []string{})
	srv.AddAddressableService(t, "db", testutil.HealthPassing,
		"10.10.10.10", 8000, []string{})
	return srv
}

func runSyncStop(configPath string, dur time.Duration) error {
	cmd, err := runSync(configPath)
	if err != nil {
		return err
	}
	time.Sleep(dur)
	return stopCommand(cmd)
}

func runSyncOnce(configPath string) error {
	cmd := exec.Command("consul-terraform-sync", "-once", fmt.Sprintf("--config-file=%s", configPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkStateFile(t *testing.T, consulAddr, taskname string) int {
	u := fmt.Sprintf("http://%s/v1/kv/%s-env:%s", consulAddr, config.DefaultTFBackendKVPath, taskname)
	resp := testutils.RequestHTTP(t, http.MethodGet, u, "")
	defer resp.Body.Close()
	return resp.StatusCode
}

// checkStateFileLocally returns whether or not a statefile exists
func checkStateFileLocally(stateFilePath string) (bool, error) {
	files, err := ioutil.ReadDir(stateFilePath)
	if err != nil {
		return false, err
	}

	if len(files) != 1 {
		return false, nil
	}

	stateFile := files[0]
	if stateFile.Name() != "terraform.tfstate" {
		return false, nil
	}

	return true, nil
}
