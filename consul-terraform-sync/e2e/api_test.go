// +build e2e

package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/consul-terraform-sync/api"
	"github.com/hashicorp/consul-terraform-sync/testutils"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_StatusEndpoints(t *testing.T) {
	t.Parallel()

	srv := newTestConsulServer(t)
	defer srv.Stop()

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "status_endpoints")
	delete := testutils.MakeTempDir(t, tempDir)
	// no defer to delete directory: only delete at end of test if no errors

	port, err := api.FreePort()
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, configFile)
	config := fakeHandlerConfig().appendPort(port).
		appendConsulBlock(srv).appendTerraformBlock(tempDir)
	config.write(t, configPath)

	cmd, err := runSyncDevMode(configPath)
	require.NoError(t, err)

	// wait to run once before registering another instance to collect another event
	time.Sleep(7 * time.Second)
	service := testutil.TestService{
		ID:      "api-2",
		Name:    "api",
		Address: "5.6.7.8",
		Port:    8080,
	}
	testutils.RegisterConsulService(t, srv, service, testutil.HealthPassing)

	// wait and then retrieve status
	time.Sleep(7 * time.Second)

	taskCases := []struct {
		name       string
		path       string
		statusCode int
		expected   map[string]api.TaskStatus
	}{
		{
			"all task statuses",
			"status/tasks",
			http.StatusOK,
			map[string]api.TaskStatus{
				fakeSuccessTaskName: api.TaskStatus{
					TaskName:  fakeSuccessTaskName,
					Status:    api.StatusSuccessful,
					Enabled:   true,
					Providers: []string{"fake-sync"},
					Services:  []string{"api"},
					EventsURL: "/v1/status/tasks/fake_handler_success_task?include=events",
				},
				fakeFailureTaskName: api.TaskStatus{
					TaskName:  fakeFailureTaskName,
					Status:    api.StatusErrored,
					Enabled:   true,
					Providers: []string{"fake-sync"},
					Services:  []string{"api"},
					EventsURL: "/v1/status/tasks/fake_handler_failure_task?include=events",
				},
				disabledTaskName: api.TaskStatus{
					TaskName:  disabledTaskName,
					Status:    api.StatusUnknown,
					Enabled:   false,
					Providers: []string{"fake-sync"},
					Services:  []string{"api"},
					EventsURL: "",
				},
			},
		},
		{
			"existing single task status",
			"status/tasks/" + fakeSuccessTaskName,
			http.StatusOK,
			map[string]api.TaskStatus{
				fakeSuccessTaskName: api.TaskStatus{
					TaskName:  fakeSuccessTaskName,
					Status:    api.StatusSuccessful,
					Enabled:   true,
					Providers: []string{"fake-sync"},
					Services:  []string{"api"},
					EventsURL: "/v1/status/tasks/fake_handler_success_task?include=events",
				},
			},
		},
		{
			"non-existing single task status",
			"status/tasks/" + "non-existing-task",
			http.StatusNotFound,
			map[string]api.TaskStatus{},
		},
	}

	for _, tc := range taskCases {
		t.Run(tc.name, func(t *testing.T) {
			u := fmt.Sprintf("http://localhost:%d/%s/%s", port, "v1", tc.path)
			resp := testutils.RequestHTTP(t, http.MethodGet, u, "")
			defer resp.Body.Close()

			assert.Equal(t, tc.statusCode, resp.StatusCode)

			if tc.statusCode != http.StatusOK {
				return
			}

			var taskStatuses map[string]api.TaskStatus
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&taskStatuses)
			require.NoError(t, err)

			// clear out some event data that we'll skip checking
			for _, stat := range taskStatuses {
				for ix, event := range stat.Events {
					event.ID = ""
					event.StartTime = time.Time{}
					event.EndTime = time.Time{}
					stat.Events[ix] = event
				}
			}

			assert.Equal(t, tc.expected, taskStatuses)
		})
	}

	eventCases := []struct {
		name               string
		path               string
		expectSuccessTask  bool
		expectFailureTask  bool
		expectDisabledTask bool
	}{
		{
			"events: all task statuses",
			"status/tasks?include=events",
			true,
			true,
			true,
		},
		{
			"events: all task statuses filtered by status",
			"status/tasks?status=errored&include=events",
			false,
			true,
			false,
		},
	}

	for _, tc := range eventCases {
		t.Run(tc.name, func(t *testing.T) {
			u := fmt.Sprintf("http://localhost:%d/%s/%s", port, "v1", tc.path)
			resp := testutils.RequestHTTP(t, http.MethodGet, u, "")
			defer resp.Body.Close()

			require.Equal(t, resp.StatusCode, http.StatusOK)

			var taskStatuses map[string]api.TaskStatus
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&taskStatuses)
			require.NoError(t, err)

			checkEvents(t, taskStatuses, fakeFailureTaskName, tc.expectFailureTask)
			checkEvents(t, taskStatuses, fakeSuccessTaskName, tc.expectSuccessTask)

			task, ok := taskStatuses[disabledTaskName]
			if tc.expectDisabledTask {
				assert.True(t, ok)
				assert.Nil(t, task.Events)
			} else {
				assert.False(t, ok)
			}
		})
	}

	overallCases := []struct {
		name string
		path string
	}{
		{
			"overall status",
			"status",
		},
	}

	for _, tc := range overallCases {
		t.Run(tc.name, func(t *testing.T) {
			u := fmt.Sprintf("http://localhost:%d/%s/%s", port, "v1", tc.path)
			resp := testutils.RequestHTTP(t, http.MethodGet, u, "")
			defer resp.Body.Close()

			require.Equal(t, resp.StatusCode, http.StatusOK)

			var overallStatus api.OverallStatus
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&overallStatus)
			require.NoError(t, err)

			// check status values
			assert.Equal(t, 1, overallStatus.TaskSummary.Status.Successful)
			assert.Equal(t, 1, overallStatus.TaskSummary.Status.Unknown)
			// failed task might be errored/critical by now depending on number of events
			assert.Equal(t, 1, overallStatus.TaskSummary.Status.Errored+
				overallStatus.TaskSummary.Status.Critical)

			// check enabled values
			assert.Equal(t, 2, overallStatus.TaskSummary.Enabled.True)
			assert.Equal(t, 1, overallStatus.TaskSummary.Enabled.False)
		})
	}

	err = stopCommand(cmd)
	require.NoError(t, err)
	delete()
}

func TestE2E_TaskEndpoints_UpdateEnableDisable(t *testing.T) {
	t.Parallel()
	// Test enabling and disabling a task
	// 1. Start with disabled task. Confirm task not initialized, resources not created
	// 2. API to inspect enabling task. Confirm plan looks good, resources not created
	// 3. API to actually enable task. Confirm resources are created
	// 4. API to disable task. Delete resources. Register new service. Confirm
	// new service registering does not trigger creating resources

	srv := newTestConsulServer(t)
	defer srv.Stop()

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "disabled_task")
	delete := testutils.MakeTempDir(t, tempDir)
	// no defer to delete directory: only delete at end of test if no errors

	port, err := api.FreePort()
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, configFile)
	config := disabledTaskConfig().appendPort(port).
		appendConsulBlock(srv).appendTerraformBlock(tempDir)
	config.write(t, configPath)

	cmd, err := runSync(configPath)
	defer stopCommand(cmd)
	require.NoError(t, err)
	time.Sleep(5 * time.Second)

	// Confirm that terraform files were not generated for a disabled task
	fileInfos, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", tempDir, "disabled_task"))
	require.NoError(t, err)
	require.Equal(t, len(fileInfos), 0)

	// Confirm that resources were not created
	resourcesPath := fmt.Sprintf("%s/%s", tempDir, resourcesDir)
	confirmDirNotFound(t, resourcesPath)

	// Update Task API: enable task with inspect run option
	baseUrl := fmt.Sprintf("http://localhost:%d/%s/tasks/%s",
		port, "v1", disabledTaskName)
	u := fmt.Sprintf("%s?run=inspect", baseUrl)
	resp := testutils.RequestHTTP(t, http.MethodPatch, u, `{"enabled":true}`)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var r api.UpdateTaskResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&r)
	require.NoError(t, err)

	// Confirm inspect plan response: changes present, plan not empty
	assert.NotNil(t, r.Inspect)
	assert.True(t, r.Inspect.ChangesPresent)
	assert.NotEmpty(t, r.Inspect.Plan)

	// Confirm that resources were not generated during inspect mode
	confirmDirNotFound(t, resourcesPath)

	// Update Task API: enable task with run now option
	u = fmt.Sprintf("%s?run=now", baseUrl)
	resp1 := testutils.RequestHTTP(t, http.MethodPatch, u, `{"enabled":true}`)
	defer resp1.Body.Close()

	// Confirm that resources are generated
	_, err = ioutil.ReadDir(resourcesPath)
	require.NoError(t, err)

	// Update Task API: disable task
	resp2 := testutils.RequestHTTP(t, http.MethodPatch, baseUrl, `{"enabled":false}`)
	defer resp2.Body.Close()

	// Delete Resources
	err = os.RemoveAll(resourcesPath)
	require.NoError(t, err)

	// Register a new service
	service := testutil.TestService{
		ID:      "api-2",
		Name:    "api",
		Address: "5.6.7.8",
		Port:    8080,
	}
	testutils.RegisterConsulService(t, srv, service, testutil.HealthPassing)
	time.Sleep(3 * time.Second)

	// Confirm that resources are not recreated for disabled task
	confirmDirNotFound(t, resourcesPath)

	delete()
}

// runSyncDevMode runs the daemon in development which does not run or download
// Terraform.
func runSyncDevMode(configPath string) (*exec.Cmd, error) {
	cmd := exec.Command("consul-terraform-sync",
		fmt.Sprintf("--config-file=%s", configPath), "--client-type=development")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func runSync(configPath string) (*exec.Cmd, error) {
	cmd := exec.Command("consul-terraform-sync",
		fmt.Sprintf("--config-file=%s", configPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func stopCommand(cmd *exec.Cmd) error {
	cmd.Process.Signal(os.Interrupt)
	sigintErr := errors.New("signal: interrupt")
	if err := cmd.Wait(); err != nil && err != sigintErr {
		return err
	}
	return nil
}

// checkEvents does some basic checks to loosely ensure returned events in
// responses are as expected
func checkEvents(t *testing.T, taskStatuses map[string]api.TaskStatus,
	taskName string, expect bool) {

	task, ok := taskStatuses[taskName]

	if expect {
		assert.True(t, ok)
	} else {
		assert.False(t, ok)
		return
	}

	// there should be 2-4 events
	msg := fmt.Sprintf("%s expected 2-4 events, got %d", taskName, len(task.Events))
	require.True(t, 2 <= len(task.Events) && len(task.Events) <= 4, msg)

	for ix, e := range task.Events {
		assert.Equal(t, taskName, e.TaskName)

		require.NotNil(t, e.Config)
		assert.Equal(t, []string{"fake-sync"}, e.Config.Providers)
		assert.Equal(t, []string{"api"}, e.Config.Services)
		assert.Equal(t, "../../test_modules/e2e_basic_task", e.Config.Source)

		if taskName == fakeSuccessTaskName {
			assert.True(t, e.Success)
		}

		if taskName == fakeFailureTaskName {
			// last event should be successful, others failure
			msg := fmt.Sprintf("Event %d of %d: %v", ix+1, len(task.Events), e)
			if ix == len(task.Events)-1 {
				assert.True(t, e.Success, msg)
			} else {
				assert.False(t, e.Success, msg)
			}
		}
	}
}

func confirmDirNotFound(t *testing.T, dir string) {
	_, err := ioutil.ReadDir(dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no such file or directory")
}
