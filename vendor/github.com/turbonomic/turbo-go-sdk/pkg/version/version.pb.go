// Code generated by protoc-gen-go.
// source: version.proto
// DO NOT EDIT!

/*
Package version_dto is a generated protocol buffer package.

It is generated from these files:
	version.proto

It has these top-level messages:
	NegotiationRequest
	NegotiationAnswer
*/
package version

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type NegotiationAnswer_NegotiationStatus int32

const (
	// Server accepts the version of client's protocol, which is fully functional.
	NegotiationAnswer_ACCEPTED NegotiationAnswer_NegotiationStatus = 0
	// Server either do not know about this version (server is obsolete)
	// or the version is no longer supported
	NegotiationAnswer_REJECTED NegotiationAnswer_NegotiationStatus = 1
	// Used for specific cases, where protocol version can be used, but some another conditions failed
	NegotiationAnswer_OTHER NegotiationAnswer_NegotiationStatus = 2
)

var NegotiationAnswer_NegotiationStatus_name = map[int32]string{
	0: "ACCEPTED",
	1: "REJECTED",
	2: "OTHER",
}
var NegotiationAnswer_NegotiationStatus_value = map[string]int32{
	"ACCEPTED": 0,
	"REJECTED": 1,
	"OTHER":    2,
}

func (x NegotiationAnswer_NegotiationStatus) Enum() *NegotiationAnswer_NegotiationStatus {
	p := new(NegotiationAnswer_NegotiationStatus)
	*p = x
	return p
}
func (x NegotiationAnswer_NegotiationStatus) String() string {
	return proto.EnumName(NegotiationAnswer_NegotiationStatus_name, int32(x))
}
func (x *NegotiationAnswer_NegotiationStatus) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(NegotiationAnswer_NegotiationStatus_value, data, "NegotiationAnswer_NegotiationStatus")
	if err != nil {
		return err
	}
	*x = NegotiationAnswer_NegotiationStatus(value)
	return nil
}
func (NegotiationAnswer_NegotiationStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{1, 0}
}

// Version negotiation request. This should be sent by client to server
type NegotiationRequest struct {
	// This is the version of protocol at the client side
	ProtocolVersion  *string `protobuf:"bytes,1,req,name=protocol_version,json=protocolVersion" json:"protocol_version,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *NegotiationRequest) Reset()                    { *m = NegotiationRequest{} }
func (m *NegotiationRequest) String() string            { return proto.CompactTextString(m) }
func (*NegotiationRequest) ProtoMessage()               {}
func (*NegotiationRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *NegotiationRequest) GetProtocolVersion() string {
	if m != nil && m.ProtocolVersion != nil {
		return *m.ProtocolVersion
	}
	return ""
}

// Negotiation result. Server answers with this message to a request by NegotiationRequest
type NegotiationAnswer struct {
	// The status, that shows, whether this certain version of protocol can be used to
	// communicate between client and server
	NegotiationResult *NegotiationAnswer_NegotiationStatus `protobuf:"varint,1,req,name=negotiation_result,json=negotiationResult,enum=version_dto.NegotiationAnswer_NegotiationStatus" json:"negotiation_result,omitempty"`
	// Description section, explaining the reasons and details of the certain status
	Description      *string `protobuf:"bytes,2,req,name=description" json:"description,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *NegotiationAnswer) Reset()                    { *m = NegotiationAnswer{} }
func (m *NegotiationAnswer) String() string            { return proto.CompactTextString(m) }
func (*NegotiationAnswer) ProtoMessage()               {}
func (*NegotiationAnswer) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *NegotiationAnswer) GetNegotiationResult() NegotiationAnswer_NegotiationStatus {
	if m != nil && m.NegotiationResult != nil {
		return *m.NegotiationResult
	}
	return NegotiationAnswer_ACCEPTED
}

func (m *NegotiationAnswer) GetDescription() string {
	if m != nil && m.Description != nil {
		return *m.Description
	}
	return ""
}

func init() {
	proto.RegisterType((*NegotiationRequest)(nil), "version_dto.NegotiationRequest")
	proto.RegisterType((*NegotiationAnswer)(nil), "version_dto.NegotiationAnswer")
	proto.RegisterEnum("version_dto.NegotiationAnswer_NegotiationStatus", NegotiationAnswer_NegotiationStatus_name, NegotiationAnswer_NegotiationStatus_value)
}

func init() { proto.RegisterFile("version.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 240 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0x4f, 0x4b, 0xc3, 0x40,
	0x10, 0x47, 0x4d, 0x40, 0xb0, 0x53, 0xff, 0xa4, 0x73, 0xea, 0xb1, 0xe6, 0xa4, 0x97, 0x45, 0xbc,
	0x08, 0x5e, 0xa4, 0x8d, 0x0b, 0xe2, 0x41, 0x65, 0x2d, 0x5e, 0x43, 0xdd, 0x2e, 0xb2, 0xd0, 0xec,
	0xd4, 0xdd, 0xd9, 0xfa, 0x3d, 0xfd, 0x44, 0x92, 0x34, 0xc1, 0x15, 0x8f, 0xbf, 0xc7, 0xe4, 0xe5,
	0xb1, 0x70, 0xb2, 0x33, 0x3e, 0x58, 0x72, 0x62, 0xeb, 0x89, 0x09, 0xc7, 0xfd, 0xac, 0xd7, 0x4c,
	0xe5, 0x1d, 0xe0, 0x93, 0xf9, 0x20, 0xb6, 0x2b, 0xb6, 0xe4, 0x94, 0xf9, 0x8c, 0x26, 0x30, 0x5e,
	0x42, 0xd1, 0xdd, 0x6a, 0xda, 0xd4, 0xfd, 0xf5, 0x34, 0x9b, 0xe5, 0x17, 0x23, 0x75, 0x36, 0xf0,
	0xb7, 0x3d, 0x2e, 0xbf, 0x33, 0x98, 0x24, 0x86, 0xb9, 0x0b, 0x5f, 0xc6, 0x63, 0x0d, 0xe8, 0x7e,
	0x61, 0xed, 0x4d, 0x88, 0x1b, 0xee, 0x14, 0xa7, 0xd7, 0x57, 0x22, 0x09, 0x10, 0xff, 0xbe, 0x4d,
	0xc9, 0x2b, 0xaf, 0x38, 0x06, 0x35, 0x71, 0x69, 0x62, 0xab, 0xc2, 0x19, 0x8c, 0xd7, 0x26, 0x68,
	0x6f, 0xb7, 0x2d, 0x9c, 0xe6, 0x5d, 0x5c, 0x8a, 0xca, 0xdb, 0x3f, 0x5d, 0x7b, 0x13, 0x1e, 0xc3,
	0xd1, 0xbc, 0xaa, 0xe4, 0xcb, 0x52, 0xde, 0x17, 0x07, 0xed, 0x52, 0xf2, 0x51, 0x56, 0xed, 0xca,
	0x70, 0x04, 0x87, 0xcf, 0xcb, 0x07, 0xa9, 0x8a, 0x7c, 0x71, 0x03, 0xe7, 0x9a, 0x1a, 0xb1, 0x6b,
	0x38, 0xfa, 0x77, 0x12, 0x9a, 0x9a, 0x26, 0x3a, 0xab, 0x3b, 0xd3, 0x50, 0xbf, 0xc0, 0xfe, 0x09,
	0x92, 0xbf, 0xfc, 0x04, 0x00, 0x00, 0xff, 0xff, 0xea, 0x71, 0xbc, 0xa6, 0x6b, 0x01, 0x00, 0x00,
}
