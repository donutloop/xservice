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

package helloworld_test

import (
	"context"
	"github.com/donutloop/xservice/integration_tests/api_hello_world"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type HelloWorldServer struct{}

func (s *HelloWorldServer) Hello(ctx context.Context, req *helloworld.HelloReq) (*helloworld.HelloResp, error) {
	return &helloworld.HelloResp{Text: "Hello " + req.Subject}, nil
}

var client helloworld.HelloWorld

func TestMain(m *testing.M) {
	handler := helloworld.NewHelloWorldServer(&HelloWorldServer{}, nil)
	mux := http.NewServeMux()
	mux.Handle(helloworld.HelloWorldPathPrefix+"Hello", handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	client = helloworld.NewHelloWorldJSONClient(server.URL, &http.Client{})

	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

func TestHelloWorldCall(t *testing.T) {
	resp, err := client.Hello(context.Background(), &helloworld.HelloReq{Subject: "World"})
	if err != nil {
		t.Fatal(err)
	}

	expectedMessage := "Hello World"
	if resp.Text != expectedMessage {
		t.Fatalf(`unexpected text (actual: "%s", expected: "%s")`, resp.Text, expectedMessage)
	}
}
