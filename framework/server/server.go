package server

import "net/http"

type Server interface {
	http.Handler
	// ServiceDescriptor returns gzipped bytes describing the .proto file that
	// this service was generated from. Once unzipped, the bytes can be
	// unmarshalled as a github.com/golang/protobuf/protoc-gen-go/descriptor.FileDescriptorProto.
	ServiceDescriptor() ([]byte, int)
	// ProtocGenTwirpVersion is the semantic version string of the version of xservice
	ProtocGenXServiceVersion() string
}
