# xservice Developer Quickstart Guide

## Requirements

* Go 1.8+
* protoc, the protobuf compiler. You need version 3+.
* github.com/golang/protobuf/protoc-gen-go, the Go protobuf generator plugin. Get this with go get.

## Godep

Use dep (https://github.com/golang/dep) to add/update dependencies.

As we don't commit vendor into our release code!

## Prepare GO development environment

Follow https://golang.org/doc/install to install golang.
Make sure you have your $GOPATH, $PATH setup correctly

## Clone xservice code

Clone or copy the code into $GOPATH/src/github.com/donutloop/xservice

## install xservice

```bash
$ go get -u github.com/golang/dep/cmd/dep

# Download deps 
dep ensure 

cd $GOPATH/src/github.com/donutloop/xservice 
$ go install -v ./...
```

## Run xservice

```bash
$ cd $PROJECT
$ protoc -I . service.proto --xservice_out=. --go_out=. 
```

## Test xservice 

```bash
$ make test
```
