
![alt text](logo.png "XService")

[![Build Status](https://travis-ci.org/donutloop/xservice.svg?branch=master)](https://travis-ci.org/donutloop/xservice)
[![Coverage Status](https://coveralls.io/repos/github/donutloop/xservice/badge.svg)](https://coveralls.io/github/donutloop/xservice)
[![Go Report Card](https://goreportcard.com/badge/github.com/donutloop/xservice)](https://goreportcard.com/report/github.com/donutloop/xservice)
## Introduction

xservice is a simple generator library used for generating services quickly and easily for an proto buffer. 
The purpose of this project is to generate of a lot of the basic boilerplate associated with writing API services so that you can focus on writing business logic.

## Making a golang xservice

To make a golang service:

  1. Define your service in a **Proto** file.
  2. Use the `protoc` command to generate go code from the **Proto** file, it
     will generate an **interface**, a **server** and some **server utils** (to
     easily start an http listener).
  3. Implement the generated **interface** to implement the service.

For example, a HelloWorld **Proto** file:

```protobuf
syntax = "proto3";
package donutloop.xservice.example.helloworld;
option go_package = "helloworld";

service HelloWorld {
  rpc Hello(HelloReq) returns (HelloResp);
}

message HelloReq {
  string subject = 1;
}

message HelloResp {
  string text = 1;
}
```

From which xservice can auto-generate this **interface** (running the `protoc` command):

```go
type HelloWorld interface {
	Hello(context.Context, *HelloReq) (*HelloResp, error)
}
```

You provide the **implementation**:

```go
package main

import (
	"context"
	"net/http"

	pb "github.com/donutloop/xservice-example/helloworld"
)

type HelloWorldServer struct{}

func (s *HelloWorldServer) Hello(ctx context.Context, req *pb.HelloReq) (*pb.HelloResp, error) {
	return &pb.HelloResp{Text: "Hello " + req.Subject}, nil
}

// Run the implementation in a local server
func main() {
	handler := pb.NewHelloWorldServer(&HelloWorldServer{}, nil)
	// You can use any mux you like - NewHelloWorldServer gives you an http.Handler.
	mux := http.NewServeMux()
	// The generated code includes a const, <ServiceName>PathPrefix, which
	// can be used to mount your service on a mux.
	mux.Handle(pb.HelloWorldPathPrefix, handler)
	http.ListenAndServe(":8080", mux)
}
```

 Now you can just use the auto-generated Client to make remote calls to your new service:

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	pb "github.com/donutloop/xservice-example/helloworld"
)

func main() {
	client := pb.NewHelloWorldJSONClient("http://localhost:8080", &http.Client{})

	resp, err := client.Hello(context.Background(), &pb.HelloReq{Subject: "World"})
	if err == nil {
		fmt.Println(resp.Text) // prints "Hello World"
	}
}
```

## QuickStart for developers

Please refer [**docs/DeveloperQuickStart.md**](https://github.com/donutloop/xservice/blob/master/docs/DeveloperQuickstartGuide.md)

## Roadmap

* Multi language support

## Contribution

Thank you for considering to help out with the source code! We welcome contributions from
anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to xservice, please fork, fix, commit and send a pull request
for the maintainers to review and merge into the main code base to ensure those changes are in line with the general philosophy of the project and/or get some
early feedback which can make both your efforts much lighter as well as our review and merge
procedures quick and simple.

Please read and follow our [Contributing](https://github.com/donutloop/xservice/blob/master/CONTRIBUTING.md).

## Code of Conduct

Please read and follow our [Code of Conduct](https://github.com/donutloop/xservice/blob/master/CODE_OF_CONDUCT.md).

## Credits

* Parts of xservice thinking comes from twirp (https://github.com/twitchtv/twirp)