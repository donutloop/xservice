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

package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/donutloop/xservice/framework/errors"
	"github.com/donutloop/xservice/framework/hooks"
	"github.com/donutloop/xservice/framework/xcontext"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

// HTTPClient is the interface used by generated clients to send HTTP requests.
// It is fulfilled by *(net/http).Client, which is sufficient for most users.
// Users can provide their own implementation for special retry policies.
//
// HTTPClient implementations should not follow redirects. Redirects are
// automatically disabled if *(net/http).Client is passed to client
// constructors. See the withoutRedirects function in this file for more
// details.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// getCustomHTTPReqHeaders retrieves a copy of any headers that are set in
// a context through the .WithHTTPRequestHeaders function.
// If there are no headers set, or if they have the wrong type, nil is returned.
func getCustomHTTPReqHeaders(ctx context.Context) http.Header {
	header, ok := xcontext.HTTPRequestHeaders(ctx)
	if !ok || header == nil {
		return nil
	}
	copied := make(http.Header)
	for k, vv := range header {
		if vv == nil {
			copied[k] = nil
			continue
		}
		copied[k] = make([]string, len(vv))
		copy(copied[k], vv)
	}
	return copied
}

// WriteError writes an HTTP response with a valid  error format.
// If err is not a .Error, it will get wrapped with .InternalErrorWith(err)
func WriteError(resp http.ResponseWriter, err error) {
	WriteErrorAndTriggerHooks(context.Background(), resp, err, nil)
}

// writeError writes  errors in the response and triggers hooks.
func WriteErrorAndTriggerHooks(ctx context.Context, resp http.ResponseWriter, err error, hooks *hooks.ServerHooks) {
	// Non- errors are wrapped as Internal (default)
	terr, ok := err.(errors.Error)
	if !ok {
		terr = errors.InternalErrorWith(err)
	}

	statusCode := errors.ServerHTTPStatusFromErrorCode(terr.Code())
	ctx = xcontext.WithStatusCode(ctx, statusCode)
	ctx = CallError(ctx, hooks, terr)

	resp.Header().Set("Content-Type", "application/json") // Error responses are always JSON (instead of protobuf)
	resp.WriteHeader(statusCode)                          // HTTP response status code

	respBody := marshalErrorToJSON(terr)
	_, err2 := resp.Write(respBody)
	if err2 != nil {
		log.Printf("unable to send error message %q: %s", terr, err2)
	}

	CallResponseSent(ctx, hooks)
}

// urlBase helps ensure that addr specifies a scheme. If it is unparsable
// as a URL, it returns addr unchanged.
func UrlBase(addr string) string {
	// If the addr specifies a scheme, use it. If not, default to
	// http. If url.Parse fails on it, return it unchanged.
	url, err := url.Parse(addr)
	if err != nil {
		return addr
	}
	if url.Scheme == "" {
		url.Scheme = "http"
	}
	return url.String()
}

// closebody closes a response or request body and just logs
// any error encountered while closing, since errors are
// considered very unusual.
func Closebody(body io.Closer, errorFunc LogErrorFunc) {
	if err := body.Close(); err != nil {
		errorFunc("error closing body %v", err)
	}
}

// newRequest makes an http.Request from a client, adding common headers.
func newRequest(ctx context.Context, url string, reqBody io.Reader, contentType string) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if customHeader := getCustomHTTPReqHeaders(ctx); customHeader != nil {
		req.Header = customHeader
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("XService-Version", "v0.1.0")
	return req, nil
}

// JSON serialization for errors
type errJSON struct {
	Code string            `json:"code"`
	Msg  string            `json:"msg"`
	Meta map[string]string `json:"meta,omitempty"`
}

// marshalErrorToJSON returns JSON from a .Error, that can be used as HTTP error response body.
// If serialization fails, it will use a descriptive Internal error instead.
func marshalErrorToJSON(terr errors.Error) []byte {
	// make sure that msg is not too large
	msg := terr.Msg()
	if len(msg) > 1e6 {
		msg = msg[:1e6]
	}

	tj := errJSON{
		Code: string(terr.Code()),
		Msg:  msg,
		Meta: terr.MetaMap(),
	}

	buf, err := json.Marshal(&tj)
	if err != nil {
		buf = []byte("{\"type\": \"" + errors.Internal + "\", \"msg\": \"There was an error but it could not be serialized into JSON\"}") // fallback
	}

	return buf
}

// doProtobufRequest is common code to make a request to the remote  service.
func doProtobufRequest(ctx context.Context, client HTTPClient, url string, in, out proto.Message) (err error) {
	reqBodyBytes, err := proto.Marshal(in)
	if err != nil {
		return errors.ClientError("failed to marshal proto request", err)
	}
	reqBody := bytes.NewBuffer(reqBodyBytes)
	if err = ctx.Err(); err != nil {
		return errors.ClientError("aborted because context was done", err)
	}

	req, err := newRequest(ctx, url, reqBody, "application/protobuf")
	if err != nil {
		return errors.ClientError("could not build request", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.ClientError("failed to do request", err)
	}

	defer func() {
		cerr := resp.Body.Close()
		if err == nil && cerr != nil {
			err = errors.ClientError("failed to close response body", cerr)
		}
	}()

	if err = ctx.Err(); err != nil {
		return errors.ClientError("aborted because context was done", err)
	}

	if resp.StatusCode != http.StatusOK {
		return errorFromResponse(resp)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.ClientError("failed to read response body", err)
	}
	if err = ctx.Err(); err != nil {
		return errors.ClientError("aborted because context was done", err)
	}

	if err = proto.Unmarshal(respBodyBytes, out); err != nil {
		return errors.ClientError("failed to unmarshal proto response", err)
	}
	return nil
}

// doJSONRequest is common code to make a request to the remote  service.
func DoJSONRequest(ctx context.Context, client HTTPClient, url string, in, out proto.Message) (err error) {

	reqBody := new(bytes.Buffer)
	marshaler := &jsonpb.Marshaler{OrigName: true}
	if err = marshaler.Marshal(reqBody, in); err != nil {
		return errors.ClientError("failed to marshal json request", err)
	}
	if err = ctx.Err(); err != nil {
		return errors.ClientError("aborted because context was done", err)
	}

	req, err := newRequest(ctx, url, reqBody, "application/json")
	if err != nil {
		return errors.ClientError("could not build request", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.ClientError("failed to do request", err)
	}

	defer func() {
		cerr := resp.Body.Close()
		if err == nil && cerr != nil {
			err = errors.ClientError("failed to close response body", cerr)
		}
	}()

	if err = ctx.Err(); err != nil {
		return errors.ClientError("aborted because context was done", err)
	}

	if resp.StatusCode != http.StatusOK {
		return errorFromResponse(resp)
	}

	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}
	if err = unmarshaler.Unmarshal(resp.Body, out); err != nil {
		return errors.ClientError("failed to unmarshal json response", err)
	}
	if err = ctx.Err(); err != nil {
		return errors.ClientError("aborted because context was done", err)
	}
	return nil
}

// The standard library will, by default, redirect requests (including POSTs) if it gets a 302 or
// 303 response, and also 301s in go1.8. It redirects by making a second request, changing the
// method to GET and removing the body. This produces very confusing error messages, so instead we
// set a redirect policy that always errors. This stops Go from executing the redirect.
//
// We have to be a little careful in case the user-provided http.Client has its own CheckRedirect
// policy - if so, we'll run through that policy first.
//
// Because this requires modifying the http.Client, we make a new copy of the client and return it.
func WithoutRedirects(in *http.Client) *http.Client {
	copy := *in
	copy.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if in.CheckRedirect != nil {
			// Run the input's redirect if it exists, in case it has side effects, but ignore any error it
			// returns, since we want to use ErrUseLastResponse.
			err := in.CheckRedirect(req, via)
			_ = err // Silly, but this makes sure generated code passes errcheck -blank, which some people use.
		}
		return http.ErrUseLastResponse
	}
	return &copy
}

// errorFromResponse builds a .Error from a non-200 HTTP response.
// If the response has a valid serialized  error, then it's returned.
// If not, the response status code is used to generate a similar
// error. See ErrorFromIntermediary for more info on intermediary errors.
func errorFromResponse(resp *http.Response) errors.Error {
	statusCode := resp.StatusCode
	statusText := http.StatusText(statusCode)

	if isHTTPRedirect(statusCode) {
		// Unexpected redirect: it must be an error from an intermediary.
		//  clients don't follow redirects automatically,  only handles
		// POST requests, redirects should only happen on GET and HEAD requests.
		location := resp.Header.Get("Location")
		return ErrorFromIntermediary(statusCode, fmt.Sprintf("unexpected HTTP status code %d %q received, Location=%q", statusCode, statusText, location), location)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.ClientError("failed to read server error response body", err)
	}
	var tj errJSON
	if err := json.Unmarshal(respBodyBytes, &tj); err != nil {
		// Invalid JSON response; it must be an error from an intermediary.
		return ErrorFromIntermediary(statusCode, fmt.Sprintf("Error from intermediary with HTTP status code %d %q", statusCode, statusText), string(respBodyBytes))
	}

	errorCode := errors.ErrorCode(tj.Code)
	if !errors.IsValidErrorCode(errorCode) {
		return errors.InternalError(fmt.Sprintf("invalid type returned from server error response: %s", tj.Code))
	}

	terr := errors.NewError(errorCode, tj.Msg)
	for k, v := range tj.Meta {
		terr = terr.WithMeta(k, v)
	}
	return terr
}

// ErrorFromIntermediary maps HTTP errors from sources to  errors.
// The mapping is similar to gRPC: https://github.com/grpc/grpc/blob/master/doc/http-grpc-status-mapping.md.
// Returned  Errors have some additional metadata for inspection.
func ErrorFromIntermediary(status int, msg string, bodyOrLocation string) errors.Error {
	var code errors.ErrorCode
	if isHTTPRedirect(status) { // 3xx
		code = errors.Internal
	} else {
		switch status {
		case http.StatusBadRequest:
			code = errors.Internal
		case http.StatusUnauthorized:
			code = errors.Unauthenticated
		case http.StatusForbidden:
			code = errors.PermissionDenied
		case http.StatusNotFound:
			code = errors.BadRoute
		case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			code = errors.Unavailable
		default: // All other codes
			code = errors.Unknown
		}
	}

	terr := errors.NewError(code, msg)
	terr = terr.WithMeta("http_error_from_intermediary", "true") // to easily know if this error was from intermediary
	terr = terr.WithMeta("status_code", strconv.Itoa(status))
	if isHTTPRedirect(status) {
		terr = terr.WithMeta("location", bodyOrLocation)
	} else {
		terr = terr.WithMeta("body", bodyOrLocation)
	}
	return terr
}

func isHTTPRedirect(status int) bool {
	return status >= 300 && status <= 399
}

// Call .ServerHooks.RequestReceived if the hook is available
func CallRequestReceived(ctx context.Context, h *hooks.ServerHooks) (context.Context, error) {
	if h == nil || h.RequestReceived == nil {
		return ctx, nil
	}
	return h.RequestReceived(ctx)
}

// Call .ServerHooks.RequestRouted if the hook is available
func CallRequestRouted(ctx context.Context, h *hooks.ServerHooks) (context.Context, error) {
	if h == nil || h.RequestRouted == nil {
		return ctx, nil
	}
	return h.RequestRouted(ctx)
}

// Call .ServerHooks.ResponsePrepared if the hook is available
func CallResponsePrepared(ctx context.Context, h *hooks.ServerHooks) context.Context {
	if h == nil || h.ResponsePrepared == nil {
		return ctx
	}
	return h.ResponsePrepared(ctx)
}

// Call .ServerHooks.ResponseSent if the hook is available
func CallResponseSent(ctx context.Context, h *hooks.ServerHooks) {
	if h == nil || h.ResponseSent == nil {
		return
	}
	h.ResponseSent(ctx)
}

// Call .ServerHooks.Error if the hook is available
func CallError(ctx context.Context, h *hooks.ServerHooks, err errors.Error) context.Context {
	if h == nil || h.Error == nil {
		return ctx
	}
	return h.Error(ctx, err)
}

// LogErrorFunc logs critical errors
type LogErrorFunc func(format string, args ...interface{})
