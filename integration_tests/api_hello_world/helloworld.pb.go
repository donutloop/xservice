// Code generated by protoc-gen-go. DO NOT EDIT.
// source: helloworld.proto

/*
Package helloworld is a generated protocol buffer package.

It is generated from these files:
	helloworld.proto

It has these top-level messages:
	HelloReq
	HelloResp
*/
package helloworld

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

type HelloReq struct {
	Subject string `protobuf:"bytes,1,opt,name=subject" json:"subject,omitempty"`
}

func (m *HelloReq) Reset()                    { *m = HelloReq{} }
func (m *HelloReq) String() string            { return proto.CompactTextString(m) }
func (*HelloReq) ProtoMessage()               {}
func (*HelloReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *HelloReq) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

type HelloResp struct {
	Text string `protobuf:"bytes,1,opt,name=text" json:"text,omitempty"`
}

func (m *HelloResp) Reset()                    { *m = HelloResp{} }
func (m *HelloResp) String() string            { return proto.CompactTextString(m) }
func (*HelloResp) ProtoMessage()               {}
func (*HelloResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *HelloResp) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func init() {
	proto.RegisterType((*HelloReq)(nil), "example.helloworld.HelloReq")
	proto.RegisterType((*HelloResp)(nil), "example.helloworld.HelloResp")
}

func init() { proto.RegisterFile("helloworld.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 143 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xc8, 0x48, 0xcd, 0xc9,
	0xc9, 0x2f, 0xcf, 0x2f, 0xca, 0x49, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x4a, 0xad,
	0x48, 0xcc, 0x2d, 0xc8, 0x49, 0xd5, 0x43, 0xc8, 0x28, 0xa9, 0x70, 0x71, 0x78, 0x80, 0x78, 0x41,
	0xa9, 0x85, 0x42, 0x12, 0x5c, 0xec, 0xc5, 0xa5, 0x49, 0x59, 0xa9, 0xc9, 0x25, 0x12, 0x8c, 0x0a,
	0x8c, 0x1a, 0x9c, 0x41, 0x30, 0xae, 0x92, 0x3c, 0x17, 0x27, 0x54, 0x55, 0x71, 0x81, 0x90, 0x10,
	0x17, 0x4b, 0x49, 0x6a, 0x05, 0x4c, 0x0d, 0x98, 0x6d, 0x14, 0xc4, 0xc5, 0x05, 0x56, 0x10, 0x0e,
	0x32, 0x54, 0xc8, 0x85, 0x8b, 0x15, 0xcc, 0x13, 0x92, 0xd1, 0xc3, 0xb4, 0x52, 0x0f, 0x66, 0x9f,
	0x94, 0x2c, 0x1e, 0xd9, 0xe2, 0x02, 0x27, 0x9e, 0x28, 0x2e, 0x84, 0x78, 0x12, 0x1b, 0xd8, 0x0f,
	0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x65, 0x34, 0xd5, 0xb9, 0xd7, 0x00, 0x00, 0x00,
}
