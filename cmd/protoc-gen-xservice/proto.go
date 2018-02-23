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
// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.  All rights reserved.
// https://github.com/twitchtv/twirp/

package main

import (
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"io"
	"io/ioutil"
	"os"

	"github.com/donutloop/xservice/generator/proto/go"
	"github.com/gogo/protobuf/proto"
	"log"
)

func main() {
	g := goproto.NewAPIGenerator()
	Main(g)
}

type Generator interface {
	Generate(in *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error)
}

func Main(g Generator) {
	req := readGenRequest(os.Stdin)
	resp, err := g.Generate(req)
	if err != nil {
		log.Fatal(err)
	}
	writeResponse(os.Stdout, resp)
}

func readGenRequest(r io.Reader) *plugin.CodeGeneratorRequest {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	req := new(plugin.CodeGeneratorRequest)
	if err = proto.Unmarshal(data, req); err != nil {
		log.Fatal(err)
	}

	if len(req.FileToGenerate) == 0 {
		log.Fatal("no files to generate")
	}

	return req
}

func writeResponse(w io.Writer, resp *plugin.CodeGeneratorResponse) {
	data, err := proto.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
