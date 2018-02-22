//// Code generated by xproto [v0.1.0], DO NOT EDIT.

//// source: [api_hello_world/helloworld.proto]

//Package [helloworld] is a generated stub package.

//This code was generated with github.com/donutloop/xservice [v0.1.0]

//It is generated from these files:

//	 [api_hello_world/helloworld.proto]

//package [helloworld]

package helloworld

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	jsonpb "github.com/golang/protobuf/jsonpb"

	"github.com/donutloop/xservice/framework/transport"

	"github.com/donutloop/xservice/framework/ctxsetters"

	"github.com/donutloop/xservice/framework/errors"

	"github.com/donutloop/xservice/framework/hooks"

	"github.com/donutloop/xservice/framework/server"

	"github.com/donutloop/xservice/framework/xhttp"
)

// //[HelloWorldPathPrefix HelloWorld] is used for all URL paths on a %!s(MISSING) server.

//Requests are always: POST [HelloWorldPathPrefix] /method

//It can be used in an HTTP mux to route requests
const HelloWorldPathPrefix string = "/xservice/example.helloworld.HelloWorld/"

// 152 bytes of a gzipped FileDescriptorProto
var xserviceFileDescriptor0 = []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x48, 0x2c, 0xc8, 0x8c, 0xcf, 0x48, 0xcd, 0xc9, 0xc9, 0x8f, 0x2f, 0xcf, 0x2f, 0xca, 0x49, 0xd1, 0x07, 0xb3, 0xc1, 0x4c, 0xbd, 0x82, 0xa2, 0xfc, 0x92, 0x7c, 0x21, 0xa1, 0xd4, 0x8a, 0xc4, 0xdc, 0x82, 0x9c, 0x54, 0x3d, 0x84, 0x8c, 0x92, 0x0a, 0x17, 0x87, 0x07, 0x88, 0x17, 0x94, 0x5a, 0x28, 0x24, 0xc1, 0xc5, 0x5e, 0x5c, 0x9a, 0x94, 0x95, 0x9a, 0x5c, 0x22, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0x19, 0x04, 0xe3, 0x2a, 0xc9, 0x73, 0x71, 0x42, 0x55, 0x15, 0x17, 0x08, 0x09, 0x71, 0xb1, 0x94, 0xa4, 0x56, 0xc0, 0xd4, 0x80, 0xd9, 0x46, 0x41, 0x5c, 0x5c, 0x60, 0x05, 0xe1, 0x20, 0x43, 0x85, 0x5c, 0xb8, 0x58, 0xc1, 0x3c, 0x21, 0x19, 0x3d, 0x4c, 0x2b, 0xf5, 0x60, 0xf6, 0x49, 0xc9, 0xe2, 0x91, 0x2d, 0x2e, 0x70, 0xe2, 0x89, 0xe2, 0x42, 0x88, 0x27, 0xb1, 0x81, 0xfd, 0x60, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0xe0, 0x09, 0x4e, 0x4a, 0xe7, 0x00, 0x00, 0x00}

type HelloWorld interface {
	Hello(ctx context.Context, req *HelloReq) (*HelloResp, error)
}

type HelloWorldJSONClient struct {
	client transport.HTTPClient
	urls   [1]string
}

func (c *HelloWorldJSONClient) Hello(ctx context.Context, in *HelloReq) (*HelloResp, error) {
	ctx = ctxsetters.WithPackageName(ctx, "example.helloworld")
	ctx = ctxsetters.WithServiceName(ctx, "HelloWorld")
	ctx = ctxsetters.WithMethodName(ctx, "Hello")
	out := new(HelloResp)
	err := transport.DoJSONRequest(ctx, c.client, c.urls[0], in, out)
	return out, err
}

type HelloWorldServer struct {
	HelloWorld
	hooks *hooks.ServerHooks
}

func (s *HelloWorldServer) writeError(ctx context.Context, resp http.ResponseWriter, err error) {
	transport.WriteErrorAndTriggerHooks(ctx, resp, err, s.hooks)
}

func (s *HelloWorldServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ctx = ctxsetters.WithPackageName(ctx, "example.helloworld")
	ctx = ctxsetters.WithServiceName(ctx, "HelloWorld")
	ctx = ctxsetters.WithResponseWriter(ctx, resp)
	var err error
	ctx, err = transport.CallRequestReceived(ctx, s.hooks)
	if err != nil {
		s.writeError(ctx, resp, err)
		return
	}
	if req.Method != http.MethodPost {
		msg := fmt.Sprintf("unsupported method %q (only POST is allowed)", req.Method)
		terr := errors.BadRouteError(msg, req.Method, req.URL.Path)
		s.writeError(ctx, resp, terr)
		return
	}

	switch req.URL.Path {
	case "/xservice/example.helloworld.HelloWorld/Hello":
		s.serveHello(ctx, resp, req)
		return

	default:
		msg := fmt.Sprintf("no handler for path %q", req.URL.Path)
		terr := errors.BadRouteError(msg, req.Method, req.URL.Path)
		s.writeError(ctx, resp, terr)
		return
	}

}

func (s *HelloWorldServer) serveHello(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	header := req.Header.Get(xhttp.ContentTypeHeader)
	i := strings.Index(header, ";")
	if i == -1 {
		i = len(header)
	}
	modifiedHeader := strings.ToLower(header[:i])
	modifiedHeader = strings.TrimSpace(modifiedHeader)
	if modifiedHeader == "application/json" {
		s.serveHelloJSON(ctx, resp, req)
	} else {
		msg := fmt.Sprintf("unexpected Content-Type: %q", header)
		terr := errors.BadRouteError(msg, req.Method, req.URL.Path)
		s.writeError(ctx, resp, terr)
	}
	return
}

func (s *HelloWorldServer) serveHelloJSON(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	var err error
	ctx = ctxsetters.WithMethodName(ctx, "Hello")
	ctx, err = transport.CallRequestRouted(ctx, s.hooks)
	if err != nil {
		s.writeError(ctx, resp, err)
		return
	}
	defer transport.Closebody(req.Body)

	reqContent := new(HelloReq)
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}
	err = unmarshaler.Unmarshal(req.Body, reqContent)
	if err != nil {
		err = errors.WrapErr(err, "failed to parse request json")
		terr := errors.InternalErrorWith(err)
		s.writeError(ctx, resp, terr)
		return
	}
	respContent := new(HelloResp)
	responseCallWrapper := func() {
		responseDeferWrapper := func() {
			r := recover()
			if r != nil {
				terr := errors.InternalError("Internal service panic")
				s.writeError(ctx, resp, terr)
				panic(r)
			}
		}
		defer responseDeferWrapper()

		respContent, err = s.Hello(ctx, reqContent)
	}
	responseCallWrapper()
	if err != nil {
		s.writeError(ctx, resp, err)
		return
	}
	if respContent == nil {
		terr := errors.InternalError("received a nil * HelloResp, and nil error while calling Hello. nil responses are not supported")
		s.writeError(ctx, resp, terr)
		return
	}
	ctx = transport.CallResponsePrepared(ctx, s.hooks)
	buff := new(bytes.Buffer)
	marshaler := &jsonpb.Marshaler{OrigName: true}
	err = marshaler.Marshal(buff, respContent)
	if err != nil {
		err = errors.WrapErr(err, "failed to marshal json response")
		terr := errors.InternalErrorWith(err)
		s.writeError(ctx, resp, terr)
		return
	}
	ctx = ctxsetters.WithStatusCode(ctx, http.StatusOK)
	req.Header.Set(xhttp.ContentTypeHeader, xhttp.ApplicationJson)
	resp.WriteHeader(http.StatusOK)
	respBytes := buff.Bytes()
	_, err = resp.Write(respBytes)
	if err != nil {
		return
	}
	transport.CallResponseSent(ctx, s.hooks)
}

func (s *HelloWorldServer) ServiceDescriptor() ([]uint8, int) {
	return xserviceFileDescriptor0, 0
}

func (s *HelloWorldServer) ProtocGenXServiceVersion() string {
	return "v0.1.0"
}

func NewHelloWorldJSONClient(addr string, client transport.HTTPClient) HelloWorld {
	URLBase := transport.UrlBase(addr)
	prefix := URLBase + HelloWorldPathPrefix
	urls := [1]string{
		prefix + "Hello",
	}
	httpClient, ok := client.(*http.Client)
	if ok == true {
		httpClient = transport.WithoutRedirects(httpClient)
		return &HelloWorldJSONClient{
			client: httpClient,
			urls:   urls,
		}
	}
	return &HelloWorldJSONClient{
		client: client,
		urls:   urls,
	}
}
func NewHelloWorldServer(svc HelloWorld, hooks *hooks.ServerHooks) server.Server {
	return &HelloWorldServer{
		HelloWorld: svc,
		hooks:      hooks,
	}
}
