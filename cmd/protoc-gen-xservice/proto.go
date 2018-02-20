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
	g := goproto.NewServerGenerator()
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
