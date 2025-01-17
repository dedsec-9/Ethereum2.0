// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/bloxapp/ssv/ibft/proto/state.proto

package proto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type State struct {
	Stage RoundState `protobuf:"varint,1,opt,name=stage,proto3,enum=proto.RoundState" json:"stage,omitempty"`
	// lambda is an instance unique identifier, much like a block hash in a blockchain
	Lambda []byte `protobuf:"bytes,2,opt,name=lambda,proto3" json:"lambda,omitempty"`
	// sequence number is an incremental number for each instance, much like a block number would be in a blockchain
	SeqNumber            uint64   `protobuf:"varint,3,opt,name=seq_number,json=seqNumber,proto3" json:"seq_number,omitempty"`
	PreviousLambda       []byte   `protobuf:"bytes,4,opt,name=previous_lambda,json=previousLambda,proto3" json:"previous_lambda,omitempty"`
	InputValue           []byte   `protobuf:"bytes,5,opt,name=input_value,json=inputValue,proto3" json:"input_value,omitempty"`
	Round                uint64   `protobuf:"varint,6,opt,name=round,proto3" json:"round,omitempty"`
	PreparedRound        uint64   `protobuf:"varint,7,opt,name=prepared_round,json=preparedRound,proto3" json:"prepared_round,omitempty"`
	PreparedValue        []byte   `protobuf:"bytes,8,opt,name=prepared_value,json=preparedValue,proto3" json:"prepared_value,omitempty"`
	ValidatorPk          []byte   `protobuf:"bytes,9,opt,name=validator_pk,json=validatorPk,proto3" json:"validator_pk,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *State) Reset()         { *m = State{} }
func (m *State) String() string { return proto.CompactTextString(m) }
func (*State) ProtoMessage()    {}
func (*State) Descriptor() ([]byte, []int) {
	return fileDescriptor_f98bd33b792b1786, []int{0}
}

func (m *State) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_State.Unmarshal(m, b)
}
func (m *State) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_State.Marshal(b, m, deterministic)
}
func (m *State) XXX_Merge(src proto.Message) {
	xxx_messageInfo_State.Merge(m, src)
}
func (m *State) XXX_Size() int {
	return xxx_messageInfo_State.Size(m)
}
func (m *State) XXX_DiscardUnknown() {
	xxx_messageInfo_State.DiscardUnknown(m)
}

var xxx_messageInfo_State proto.InternalMessageInfo

func (m *State) GetStage() RoundState {
	if m != nil {
		return m.Stage
	}
	return RoundState_NotStarted
}

func (m *State) GetLambda() []byte {
	if m != nil {
		return m.Lambda
	}
	return nil
}

func (m *State) GetSeqNumber() uint64 {
	if m != nil {
		return m.SeqNumber
	}
	return 0
}

func (m *State) GetPreviousLambda() []byte {
	if m != nil {
		return m.PreviousLambda
	}
	return nil
}

func (m *State) GetInputValue() []byte {
	if m != nil {
		return m.InputValue
	}
	return nil
}

func (m *State) GetRound() uint64 {
	if m != nil {
		return m.Round
	}
	return 0
}

func (m *State) GetPreparedRound() uint64 {
	if m != nil {
		return m.PreparedRound
	}
	return 0
}

func (m *State) GetPreparedValue() []byte {
	if m != nil {
		return m.PreparedValue
	}
	return nil
}

func (m *State) GetValidatorPk() []byte {
	if m != nil {
		return m.ValidatorPk
	}
	return nil
}

func init() {
	proto.RegisterType((*State)(nil), "proto.State")
}

func init() {
	proto.RegisterFile("github.com/bloxapp/ssv/ibft/proto/state.proto", fileDescriptor_f98bd33b792b1786)
}

var fileDescriptor_f98bd33b792b1786 = []byte{
	// 283 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x90, 0x4f, 0x4b, 0xc3, 0x40,
	0x10, 0xc5, 0x49, 0x6d, 0xa2, 0x9d, 0xd6, 0x8a, 0x8b, 0xc8, 0x22, 0x88, 0xa9, 0x22, 0xcd, 0x41,
	0x13, 0xd0, 0x6f, 0xe0, 0x59, 0x44, 0x22, 0x78, 0xf0, 0x12, 0x76, 0xcd, 0x1a, 0x43, 0x93, 0xec,
	0x76, 0xff, 0x04, 0x3f, 0x9a, 0x1f, 0x4f, 0x32, 0xdb, 0xf6, 0xe2, 0xa1, 0xa7, 0xc9, 0xbc, 0xf9,
	0xbd, 0xf7, 0xc2, 0xc2, 0x7d, 0x55, 0xdb, 0x6f, 0xc7, 0xd3, 0x4f, 0xd9, 0x66, 0xbc, 0x91, 0x3f,
	0x4c, 0xa9, 0xcc, 0x98, 0x3e, 0xab, 0xf9, 0x97, 0xcd, 0x94, 0x96, 0x56, 0x66, 0xc6, 0x32, 0x2b,
	0x52, 0xfc, 0x26, 0x21, 0x8e, 0x8b, 0xbb, 0xfd, 0xae, 0xd6, 0x54, 0xc6, 0x9b, 0xae, 0x7f, 0x47,
	0x10, 0xbe, 0x0d, 0x21, 0x64, 0x09, 0xa1, 0xb1, 0xac, 0x12, 0x34, 0x88, 0x83, 0x64, 0xfe, 0x70,
	0xea, 0x81, 0x34, 0x97, 0xae, 0x2b, 0x91, 0xc8, 0xfd, 0x9d, 0x9c, 0x43, 0xd4, 0xb0, 0x96, 0x97,
	0x8c, 0x8e, 0xe2, 0x20, 0x99, 0xe5, 0x9b, 0x8d, 0x5c, 0x02, 0x18, 0xb1, 0x2e, 0x3a, 0xd7, 0x72,
	0xa1, 0xe9, 0x41, 0x1c, 0x24, 0xe3, 0x7c, 0x62, 0xc4, 0xfa, 0x05, 0x05, 0xb2, 0x84, 0x13, 0xa5,
	0x45, 0x5f, 0x4b, 0x67, 0x8a, 0x8d, 0x7f, 0x8c, 0xfe, 0xf9, 0x56, 0x7e, 0xf6, 0x39, 0x57, 0x30,
	0xad, 0x3b, 0xe5, 0x6c, 0xd1, 0xb3, 0xc6, 0x09, 0x1a, 0x22, 0x04, 0x28, 0xbd, 0x0f, 0x0a, 0x39,
	0x83, 0x50, 0x0f, 0x7f, 0x45, 0x23, 0xec, 0xf0, 0x0b, 0xb9, 0x85, 0x21, 0x48, 0x31, 0x2d, 0xca,
	0xc2, 0x9f, 0x0f, 0xf1, 0x7c, 0xbc, 0x55, 0xf3, 0x7f, 0x98, 0x2f, 0x38, 0xc2, 0x82, 0x1d, 0xe6,
	0x3b, 0x16, 0x30, 0xeb, 0x59, 0x53, 0x97, 0xcc, 0x4a, 0x5d, 0xa8, 0x15, 0x9d, 0x20, 0x34, 0xdd,
	0x69, 0xaf, 0xab, 0xa7, 0x9b, 0x8f, 0xc5, 0xde, 0xa7, 0xe6, 0x11, 0x8e, 0xc7, 0xbf, 0x00, 0x00,
	0x00, 0xff, 0xff, 0xc8, 0xf5, 0xc3, 0xdf, 0xcc, 0x01, 0x00, 0x00,
}
