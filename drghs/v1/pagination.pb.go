// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pagination.proto

package drghs_v1

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type PageToken struct {
	// The offset where the next request should start.
	Offset int32 `protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	// The time when the first page request was received.
	FirstRequestTimeUsec *timestamp.Timestamp `protobuf:"bytes,2,opt,name=first_request_time_usec,json=firstRequestTimeUsec,proto3" json:"first_request_time_usec,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *PageToken) Reset()         { *m = PageToken{} }
func (m *PageToken) String() string { return proto.CompactTextString(m) }
func (*PageToken) ProtoMessage()    {}
func (*PageToken) Descriptor() ([]byte, []int) {
	return fileDescriptor_567bfb3a87c868dd, []int{0}
}

func (m *PageToken) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PageToken.Unmarshal(m, b)
}
func (m *PageToken) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PageToken.Marshal(b, m, deterministic)
}
func (m *PageToken) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PageToken.Merge(m, src)
}
func (m *PageToken) XXX_Size() int {
	return xxx_messageInfo_PageToken.Size(m)
}
func (m *PageToken) XXX_DiscardUnknown() {
	xxx_messageInfo_PageToken.DiscardUnknown(m)
}

var xxx_messageInfo_PageToken proto.InternalMessageInfo

func (m *PageToken) GetOffset() int32 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *PageToken) GetFirstRequestTimeUsec() *timestamp.Timestamp {
	if m != nil {
		return m.FirstRequestTimeUsec
	}
	return nil
}

func init() {
	proto.RegisterType((*PageToken)(nil), "drghs.v1.PageToken")
}

func init() {
	proto.RegisterFile("pagination.proto", fileDescriptor_567bfb3a87c868dd)
}

var fileDescriptor_567bfb3a87c868dd = []byte{
	// 167 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x34, 0x8d, 0x41, 0x0b, 0x82, 0x40,
	0x10, 0x46, 0x31, 0x48, 0x6a, 0xbb, 0x84, 0x44, 0x89, 0x97, 0xa4, 0x93, 0xa7, 0x95, 0xea, 0x8f,
	0x94, 0xd8, 0x59, 0x56, 0x9b, 0xdd, 0x96, 0xd2, 0xb1, 0x9d, 0xd1, 0xdf, 0x1f, 0x6a, 0x1e, 0xe7,
	0xe3, 0xbd, 0x79, 0x62, 0xdb, 0x2a, 0x63, 0x1b, 0xc5, 0x16, 0x1b, 0xd9, 0x3a, 0x64, 0x0c, 0x56,
	0x4f, 0x67, 0x5e, 0x24, 0xfb, 0x73, 0x74, 0x34, 0x88, 0xe6, 0x03, 0xe9, 0xb8, 0x97, 0x9d, 0x4e,
	0xd9, 0xd6, 0x40, 0xac, 0xea, 0x76, 0x42, 0x4f, 0xbd, 0x58, 0xdf, 0x94, 0x81, 0x1c, 0xdf, 0xd0,
	0x04, 0x7b, 0xe1, 0xa3, 0xd6, 0x04, 0x1c, 0x7a, 0xb1, 0x97, 0x2c, 0xb3, 0xff, 0x15, 0xdc, 0xc5,
	0x41, 0x5b, 0x47, 0x5c, 0x38, 0xf8, 0x76, 0x40, 0x5c, 0x0c, 0x5f, 0x8a, 0x8e, 0xa0, 0x0a, 0x17,
	0xb1, 0x97, 0x6c, 0x2e, 0x91, 0x9c, 0x3a, 0x72, 0xee, 0xc8, 0x7c, 0xee, 0x64, 0xbb, 0x51, 0xcd,
	0x26, 0x73, 0x98, 0x1f, 0x04, 0x55, 0xe9, 0x8f, 0xe4, 0xf5, 0x17, 0x00, 0x00, 0xff, 0xff, 0x60,
	0x54, 0x2d, 0x03, 0xbd, 0x00, 0x00, 0x00,
}
