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
