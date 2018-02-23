// Copyright 2018 XService, All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// This file contains some code from  https://github.com/twitchtv/twirp/:
// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
// https://github.com/twitchtv/twirp/

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
