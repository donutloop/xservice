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

package xcontext

import (
	"context"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

type contextKey int

const (
	MethodNameKey contextKey = 1 + iota
	ServiceNameKey
	PackageNameKey
	StatusCodeKey
	RequestHeaderKey
	ResponseWriterKey
)

func WithMethodName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, MethodNameKey, name)
}

func WithServiceName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, ServiceNameKey, name)
}

func WithPackageName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, PackageNameKey, name)
}

func WithStatusCode(ctx context.Context, code int) context.Context {
	return context.WithValue(ctx, StatusCodeKey, strconv.Itoa(code))
}

func WithResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, ResponseWriterKey, w)
}

// MethodName extracts the name of the method being handled in the given
// context. If it is not known, it returns ("", false).
func MethodName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(MethodNameKey).(string)
	return name, ok
}

// ServiceName extracts the name of the service handling the given context. If
// it is not known, it returns ("", false).
func ServiceName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(ServiceNameKey).(string)
	return name, ok
}

// PackageName extracts the fully-qualified protobuf package name of the service
// handling the given context. If it is not known, it returns ("", false). If
// the service comes from a proto file that does not declare a package name, it
// returns ("", true).
//
// Note that the protobuf package name can be very different than the go package
// name; the two are unrelated.
func PackageName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(PackageNameKey).(string)
	return name, ok
}

// StatusCode retrieves the status code of the response (as string like "200").
// If it is known returns (status, true).
// If it is not known, it returns ("", false).
func StatusCode(ctx context.Context) (string, bool) {
	code, ok := ctx.Value(StatusCodeKey).(string)
	return code, ok
}

// WithHTTPRequestHeaders stores an http.Header in a context.Context. When
// using a generated client, you can pass the returned context
// into any of the request methods, and the stored header will be
// included in outbound HTTP requests.
//
// This can be used to set custom HTTP headers like authorization tokens or
// client IDs. But note that HTTP headers are a implementation detail,
// only visible by middleware, not by the server implementation.
//
// WithHTTPRequestHeaders returns an error if the provided http.Header
// would overwrite a header that is needed by, like "Content-Type".
func WithHTTPRequestHeaders(ctx context.Context, h http.Header) (context.Context, error) {
	if _, ok := h["Content-Type"]; ok {
		return nil, errors.New("provided header cannot set Content-Type")
	}
	if _, ok := h["XService-Version"]; ok {
		return nil, errors.New("provided header cannot set Xservice-Version")
	}

	copied := make(http.Header, len(h))
	for k, vv := range h {
		if vv == nil {
			copied[k] = nil
			continue
		}
		copied[k] = make([]string, len(vv))
		copy(copied[k], vv)
	}

	return context.WithValue(ctx, RequestHeaderKey, copied), nil
}

func HTTPRequestHeaders(ctx context.Context) (http.Header, bool) {
	h, ok := ctx.Value(RequestHeaderKey).(http.Header)
	return h, ok
}

// SetHTTPResponseHeader sets an HTTP header key-value pair using a context
// provided by a generated server, or a child of that context.
// The server will include the header in its response for that request context.
//
// This can be used to respond with custom HTTP headers like "Cache-Control".
// But note that HTTP headers are a implementation detail,
// only visible by middleware, not by the clients or their responses.
//
// The header will be ignored (noop) if the context is invalid (i.e. using a new
// context.Background() instead of passing the context from the handler).
//
// If called multiple times with the same key, it replaces any existing values
// associated with that key.
//
// SetHTTPResponseHeader returns an error if the provided header key
// would overwrite a header that is needed by xservice, like "Content-Type".
func SetHTTPResponseHeader(ctx context.Context, key, value string) error {
	if key == "Content-Type" {
		return errors.New("header key can not be Content-Type")
	}

	responseWriter, ok := ctx.Value(ResponseWriterKey).(http.ResponseWriter)
	if ok {
		responseWriter.Header().Set(key, value)
	} // invalid context is ignored, not an error, this is to allow easy unit testing with mock servers

	return nil
}
