# xservice Developer Quickstart Guide

## Requirements

* Go 1.7+
* Docker version: >= 17.05.0-ce

## Godep

Use dep (https://github.com/golang/dep) to add/update dependencies.

As we don't commit vendor into our release code!

## Prepare GO development environment

Follow https://golang.org/doc/install to install golang.
Make sure you have your $GOPATH, $PATH setup correctly

## Clone rmd code

Clone or copy the code into $GOPATH/src/github.com/donutloop/xservice

## Build & install rmd

```
$ go get -u github.com/golang/dep/cmd/dep

# Download deps 
dep ensure 

# install xservice into $GOPATH/bin
$ go build && mv ./generate $GOPATH/bin
```

## Run xservice

```
$ $GOPATH/bin/generate --help
$ $GOPATH/bin/generate
```

## Testing Requirements

We use docker to test our auto-generated code in an isolated environment
to verify that our changes didn't introduce a couple of new bugs

## Docker Install instructions:
[Docker](https://docs.docker.com/engine/installation/)  

## Test environment 

The following command builds a container and executes the test enviroment

```bash
    docker build -t "xservice:dockerfile" -f ./Dockerfile.web .
```
