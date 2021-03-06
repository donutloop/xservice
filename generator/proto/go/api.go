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

package goproto

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/donutloop/xservice/internal/xgenerator/types"
	"github.com/donutloop/xservice/internal/xproto"
	"github.com/donutloop/xservice/internal/xproto/typesmap"
	"github.com/donutloop/xservice/internal/xproto/xprotoutil"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pkg/errors"
	"go/token"
	"strconv"
	"strings"
)

const (
	ServeJSON        string = "JSON"
	ServeProtobuffer string = "Protobuffer"
)

type API struct {
	filesHandled   int
	currentPackage string // Go name of current package we're working on

	reg *typemap.Registry

	// Map to record whether we've built each package
	pkgs          map[string]string
	pkgNamesInUse map[string]bool

	// Package naming:
	genPkgName          string // Name of the package that we're generating
	fileToGoPackageName map[*descriptor.FileDescriptorProto]string

	// List of files that were inputs to the generator. We need to hold this in
	// the struct so we can write a header for the file that lists its inputs.
	genFiles []*descriptor.FileDescriptorProto

	// Output buffer that holds the bytes we want to write out for a single file.
	// Gets reset after working on a file.
	output *bytes.Buffer
}

func NewAPIGenerator() *API {
	gen := &API{
		pkgs:                make(map[string]string),
		pkgNamesInUse:       make(map[string]bool),
		fileToGoPackageName: make(map[*descriptor.FileDescriptorProto]string),
	}
	return gen
}

func FilesToGenerate(req *plugin.CodeGeneratorRequest) []*descriptor.FileDescriptorProto {
	genFiles := make([]*descriptor.FileDescriptorProto, 0)
Outer:
	for _, name := range req.FileToGenerate {
		for _, f := range req.ProtoFile {
			if f.GetName() == name {
				genFiles = append(genFiles, f)
				continue Outer
			}
		}

	}

	return genFiles
}

func (a *API) Generate(in *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	a.genFiles = FilesToGenerate(in)

	// Collect information on types.
	a.reg = typemap.New(in.ProtoFile)

	// Register names of packages that we import.

	// Time to figure out package names of objects defined in protobuf. First,
	// we'll figure out the name for the package we're generating.
	genPkgName, err := deduceGenPkgName(a.genFiles)
	if err != nil {
		return nil, errors.Wrap(err, "todo")
	}
	a.genPkgName = genPkgName

	// Next, we need to pick names for all the files that are dependencies.
	for _, f := range in.ProtoFile {
		if fileDescSliceContains(a.genFiles, f) {
			// This is a file we are generating. It gets the shared package name.
			a.fileToGoPackageName[f] = a.genPkgName
		} else {
			// This is a dependency. Use its package name.
			name := f.GetPackage()
			if name == "" {
				name = types.BaseName(f.GetName())
			}
			name = types.Identifier(name)
			a.fileToGoPackageName[f] = name
		}
	}

	// Showtime! Generate the response.
	resp := new(plugin.CodeGeneratorResponse)
	for _, f := range a.genFiles {
		respFile, err := a.generate(f)
		if err != nil {
			return nil, err
		}
		if respFile != nil {
			resp.File = append(resp.File, respFile)
		}
	}
	return resp, nil
}

func (a *API) generate(fileDescriptor *descriptor.FileDescriptorProto) (*plugin.CodeGeneratorResponse_File, error) {
	resp := new(plugin.CodeGeneratorResponse_File)
	if len(fileDescriptor.Service) == 0 {
		return nil, nil
	}

	packageName := fileDescriptor.GetOptions().GetGoPackage()
	if packageName == "" {
		return nil, errors.New("go package property is empty")
	}

	goFile, err := types.NewGoFile(packageName, *fileDescriptor.Name)
	if err != nil {
		return nil, err
	}

	a.generateFileHeader(fileDescriptor, goFile)

	a.generateAdditionalImports(fileDescriptor, goFile)

	// For each service, generate client stubs and server
	for i, service := range fileDescriptor.Service {
		goFile, err = a.generateService(fileDescriptor, service, goFile, i)
		if err != nil {
			return nil, err
		}
	}

	goFile, err = a.generateFileDescriptor(fileDescriptor, goFile)
	if err != nil {
		return nil, err
	}

	resp.Name = proto.String(goFile.GetFileName())

	content, err := goFile.RenderAndFormatCode()
	if err != nil {
		return nil, err
	}

	resp.Content = proto.String(string(content))

	a.filesHandled++
	return resp, nil
}

func (a *API) generateFileHeader(file *descriptor.FileDescriptorProto, goFile *types.FileGenerator) (*types.FileGenerator, error) {

	c := types.NewGoComment()

	c.Pf("Code generated by xproto %s, DO NOT EDIT.", xproto.Version)
	c.Pf("source: %s ", file.GetName())
	if a.filesHandled == 0 {
		c.Pf("Package %s is a generated stub package.", a.genPkgName)
		c.Pf("This code was generated with github.com/donutloop/xservice %s", xproto.Version)
		comment, err := a.reg.FileComments(file)
		if err == nil && comment.Leading != "" {
			for _, line := range strings.Split(comment.Leading, "\n") {
				line = strings.TrimPrefix(line, " ")
				// ensure we don't escape from the block comment
				line = strings.Replace(line, "*/", "* /", -1)
				c.P(line)
			}
		}
		c.P("It is generated from these files:")
		for _, f := range a.genFiles {
			c.Pf("\t %s", f.GetName())
		}
	}
	c.Pf(`package %s`, a.genPkgName)

	if err := goFile.HeaderComment(c); err != nil {
		return nil, err
	}

	return goFile, nil
}

// generateAdditionalImports generate additional imports that are not in the standard lib from golang
func (a *API) generateAdditionalImports(file *descriptor.FileDescriptorProto, goFile *types.FileGenerator) {

	if len(file.Service) == 0 {
		return
	}

	goFile.Import("", "github.com/donutloop/xservice/framework/transport")
	goFile.Import("", "github.com/donutloop/xservice/framework/xcontext")
	goFile.Import("", "github.com/donutloop/xservice/framework/errors")
	goFile.Import("", "github.com/donutloop/xservice/framework/hooks")
	goFile.Import("", "github.com/donutloop/xservice/framework/server")
	goFile.Import("", "github.com/donutloop/xservice/framework/xhttp")
}

func (a *API) generateService(fileDescriptor *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, goFile *types.FileGenerator, index int) (*types.FileGenerator, error) {

	var err error

	// interface
	goFile, err = a.generatServiceInterface(fileDescriptor, service, goFile)
	if err != nil {
		return nil, err
	}

	// JSON Client
	goFile, err = a.generateClient(ServeJSON, fileDescriptor, service, goFile)
	if err != nil {
		return nil, err
	}

	// Protobuffer  Client
	goFile, err = a.generateClient(ServeProtobuffer, fileDescriptor, service, goFile)
	if err != nil {
		return nil, err
	}

	// Server
	goFile, err = a.generateServer(fileDescriptor, service, goFile)
	if err != nil {
		return nil, err
	}

	return goFile, nil
}

func (a *API) generatServiceInterface(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, goFile *types.FileGenerator) (*types.FileGenerator, error) {

	serviceInterface, err := types.NewGoInterface(serviceName(service))
	if err != nil {
		return nil, err
	}

	comments, err := a.reg.ServiceComments(file, service)
	if err == nil {
		comment, err := prepareComment(comments)
		if err != nil {
			if err != EmptyComment {
				return nil, err
			}
		} else {
			serviceInterface.InterfaceMetadata.HeaderComment = comment
		}
	}

	for _, method := range service.Method {
		comments, err = a.reg.MethodComments(file, service, method)
		var comment string
		var err error
		if err == nil {
			comment, err = prepareComment(comments)
			if err != nil {
				if err != EmptyComment {
					return nil, err
				}
			}
		}

		inputType, err := a.goTypeName(method.GetInputType())
		if err != nil {
			return nil, err
		}

		outputType, err := a.goTypeName(method.GetOutputType())
		if err != nil {
			return nil, err
		}

		err = serviceInterface.Prototype(methodName(method), []*types.Parameter{
			{
				NameOfParameter: "ctx",
				Typ:             types.NewUnsafeTypeReference("context.Context"),
			},
			{
				NameOfParameter: "req",
				Typ:             types.NewUnsafeTypeReference(fmt.Sprintf("*%s", inputType)),
			},
		},
			[]types.TypeReference{
				types.NewUnsafeTypeReference(fmt.Sprintf("*%s", outputType)),
				types.NewUnsafeTypeReference("error"),
			}, comment)

		if err != nil {
			return nil, err
		}
	}

	if err := goFile.Interface(serviceInterface); err != nil {
		return nil, err
	}

	return goFile, nil
}

func (a *API) generateClient(name string, fileDescriptor *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, goFile *types.FileGenerator) (*types.FileGenerator, error) {
	servName := serviceName(service)
	structName := unexported(servName) + name + "Client"
	newClientFunc := "New" + servName + name + "Client"

	methCnt := strconv.Itoa(len(service.Method))

	// Server implementation.
	structGenerator, err := types.NewGoStruct(structName, true, false)
	if err != nil {
		return nil, err
	}
	structGenerator.StructMetaData.Comment = append(structGenerator.StructMetaData.Comment, fmt.Sprintf("%s wraps an http.client and sends %s objects", structName, name))

	structGenerator.AddUnexportedField("client", types.NewUnsafeTypeReference("transport.HTTPClient"), "")
	structGenerator.AddUnexportedField("urls", types.NewUnsafeTypeReference(fmt.Sprintf("[%s]string", methCnt)), "")

	goFile, err = a.generateClientConstructor(newClientFunc, structGenerator.StructMetaData.Name, service, goFile)
	if err != nil {
		return nil, err
	}

	structGenerator, err = a.generateClientEndpoints(name, fileDescriptor, service, structGenerator)
	if err != nil {
		return nil, err
	}

	if err := goFile.TypesWithMethods(structGenerator); err != nil {
		return nil, err
	}

	return goFile, nil
}

func (a *API) generateClientConstructor(newClientFuncName, structName string, service *descriptor.ServiceDescriptorProto, goFile *types.FileGenerator) (*types.FileGenerator, error) {

	pathPrefixConst := serviceName(service) + "PathPrefix"

	comment := fmt.Sprintf("%s constructs a new client, which wraps the http.client and implements %s", newClientFuncName, serviceName(service))
	f, err := types.NewGoFunc(newClientFuncName, []*types.Parameter{
		{
			NameOfParameter: "addr",
			Typ:             types.String,
		},
		{
			NameOfParameter: "client",
			Typ:             types.NewUnsafeTypeReference("transport.HTTPClient"),
		},
	},
		[]types.TypeReference{
			types.NewUnsafeTypeReference(serviceName(service)),
		}, comment)
	if err != nil {
		return nil, err
	}

	f.DefAssginCall([]string{"URLBase"}, types.NewUnsafeTypeReference("transport.UrlBase"), []string{"addr"})
	f.DefOperation("prefix", "URLBase", token.ADD, pathPrefixConst)

	urlsSlice, err := types.NewGoSliceLiteral("urls", types.String, len(service.Method))
	if err != nil {
		return nil, err
	}
	for _, method := range service.Method {
		urlsSlice.Append(fmt.Sprintf(`prefix + "%s"`, methodName(method)))
	}

	if err := f.SliceLiteral(*urlsSlice); err != nil {
		return nil, err
	}

	f.DefAssert([]string{"httpClient", "ok"}, "client", types.NewUnsafeTypeReference("*http.Client"))
	f.DefIfBegin("ok", token.EQL, "true")

	initStructGeneratorForHttpClient, err := types.NewInitGoStruct(structName)
	if err != nil {
		return nil, err
	}

	f.DefCall([]string{"httpClient"}, types.NewUnsafeTypeReference("transport.WithoutRedirects"), []string{"httpClient"})
	initStructGeneratorForHttpClient.AddUnexportedValueToField("client", "httpClient")
	initStructGeneratorForHttpClient.AddUnexportedValueToField("urls", "urls")

	if err := f.InitStruct("return", initStructGeneratorForHttpClient, true); err != nil {
		return nil, err
	}

	f.CloseIf()

	initStructGenerator, err := types.NewInitGoStruct(structName)
	if err != nil {
		return nil, err
	}

	initStructGenerator.AddUnexportedValueToField("client", "client")
	initStructGenerator.AddUnexportedValueToField("urls", "urls")

	if err := f.InitStruct("return", initStructGenerator, true); err != nil {
		return nil, err
	}

	if err := goFile.Func(f); err != nil {
		return nil, err
	}

	return goFile, nil
}

func (a *API) generateClientEndpoints(contentType string, fileDescriptor *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, structGenerator *types.StructGenerator) (*types.StructGenerator, error) {

	for i, method := range service.Method {
		methName := methodName(method)
		pkgName := pkgName(fileDescriptor)
		servName := serviceName(service)
		inputType, err := a.goTypeName(method.GetInputType())
		if err != nil {
			return nil, err
		}
		outputType, err := a.goTypeName(method.GetOutputType())
		if err != nil {
			return nil, err
		}

		comment := fmt.Sprintf("%s sends an %s %s object to the server", methName, inputType, contentType)
		method, err := types.NewGoMethod("c", fmt.Sprintf("*%s", structGenerator.StructMetaData.Name), methName, []*types.Parameter{
			{
				NameOfParameter: "ctx",
				Typ:             types.NewUnsafeTypeReference("context.Context"),
			},
			{
				NameOfParameter: "in",
				Typ:             types.NewUnsafeTypeReference(fmt.Sprintf("*%s", inputType)),
			},
		}, []types.TypeReference{
			types.NewUnsafeTypeReference(fmt.Sprintf("*%s", outputType)),
			types.NewUnsafeTypeReference("error"),
		}, comment)

		if err != nil {
			return nil, err
		}

		method.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("xcontext.WithPackageName"), []string{"ctx", `"` + pkgName + `"`})
		method.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("xcontext.WithServiceName"), []string{"ctx", `"` + servName + `"`})
		method.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("xcontext.WithMethodName"), []string{"ctx", `"` + methName + `"`})
		method.DefNew("out", types.NewUnsafeTypeReference(outputType))
		method.DefAssginCall([]string{"err"}, types.NewUnsafeTypeReference(fmt.Sprintf("transport.Do%sRequest", contentType)), []string{"ctx", "c.client", fmt.Sprintf("c.urls[%s]", strconv.Itoa(i)), "in", "out"})
		method.Return([]string{"out", "err"})
		structGenerator.AddMethod(method)
	}

	return structGenerator, nil
}

func (a *API) generateServer(fileDescriptor *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, goFile *types.FileGenerator) (*types.FileGenerator, error) {
	// Server implementation.
	structGenerator, err := types.NewGoStruct(serviceStruct(service), true, false)
	if err != nil {
		return nil, err
	}

	structGenerator.StructMetaData.Comment = append(structGenerator.StructMetaData.Comment, fmt.Sprintf("%s wraps an endpoint and implements http.Handler.", serviceStruct(service)))

	structGenerator.Type(types.NewUnsafeTypeReference(serviceName(service)), "")
	structGenerator.AddUnexportedField("hooks", types.NewUnsafeTypeReference("*hooks.ServerHooks"), "")
	structGenerator.AddUnexportedField("logErrorFunc", types.NewUnsafeTypeReference("transport.LogErrorFunc"), "")

	goFile, err = a.generateServerConstructor(serviceName(service), structGenerator.StructMetaData.Name, goFile)
	if err != nil {
		return nil, err
	}

	structGenerator, err = a.generateServerWriteError(structGenerator)
	if err != nil {
		return nil, err
	}

	structGenerator, err = a.generateServerRouting(fileDescriptor, service, structGenerator, goFile)
	if err != nil {
		return nil, err
	}

	// Methods.
	for _, method := range service.Method {
		structGenerator, err = a.generateServerMethod(service, method, structGenerator)
		if err != nil {
			return nil, err
		}
	}

	structGenerator, err = a.generateServiceMetadataAccessors(fileDescriptor, service, structGenerator)
	if err != nil {
		return nil, err
	}

	if err := goFile.TypesWithMethods(structGenerator); err != nil {
		return nil, err
	}

	return goFile, nil
}

func (a *API) generateServerConstructor(serverName, serverStructName string, goFile *types.FileGenerator) (*types.FileGenerator, error) {

	constructorName := fmt.Sprintf("New%sServer", serverName)

	comment := fmt.Sprintf("%s constructs a new server, and implements %s", constructorName, serverName)
	f, err := types.NewGoFunc(constructorName, []*types.Parameter{
		{
			NameOfParameter: "svc",
			Typ:             types.NewUnsafeTypeReference(serverName),
		},
		{
			NameOfParameter: "hooks",
			Typ:             types.NewUnsafeTypeReference("*hooks.ServerHooks"),
		},
		{
			NameOfParameter: "errorFunc",
			Typ:             types.NewUnsafeTypeReference("...transport.LogErrorFunc"),
		},
	},
		[]types.TypeReference{
			types.NewUnsafeTypeReference("server.Server"),
		}, comment)
	if err != nil {
		return nil, err
	}

	initStructGenerator, err := types.NewInitGoStruct(serverStructName)
	if err != nil {
		return nil, err
	}

	initStructGenerator.AddExportedValueToField(serverName, "svc")
	initStructGenerator.AddUnexportedValueToField("hooks", "hooks")

	if err := f.InitStruct("server :=", initStructGenerator, true); err != nil {
		return nil, err
	}

	f.DefIfBegin("len(errorFunc)", token.EQL, "1")
	f.StructAssignment("server", "logErrorFunc", "errorFunc[0]")
	f.Else()
	f.StructAssignment("server", "logErrorFunc", "log.Printf")
	f.CloseIf()
	f.Return([]string{"server"})

	if err := goFile.Func(f); err != nil {
		return nil, err
	}

	return goFile, nil
}

func (a *API) generateServerWriteError(structGenerator *types.StructGenerator) (*types.StructGenerator, error) {
	method, err := types.NewGoMethod("s", fmt.Sprintf("*%s", structGenerator.StructMetaData.Name), "writeError", []*types.Parameter{
		{
			NameOfParameter: "ctx",
			Typ:             types.NewUnsafeTypeReference("context.Context"),
		},
		{
			NameOfParameter: "resp",
			Typ:             types.NewUnsafeTypeReference("http.ResponseWriter"),
		},
		{
			NameOfParameter: "err",
			Typ:             types.NewUnsafeTypeReference("error"),
		},
	}, nil, "")

	if err != nil {
		return nil, err
	}
	// todo error handling
	method.Caller(types.NewUnsafeTypeReference("transport.WriteErrorAndTriggerHooks"), []string{"ctx", "resp", "err", "s.hooks"})
	structGenerator.AddMethod(method)
	return structGenerator, nil
}

func (a *API) generateServerRouting(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, structGenerator *types.StructGenerator, goFile *types.FileGenerator) (*types.StructGenerator, error) {

	pkgName := pkgName(file)
	servName := serviceName(service)
	pathPrefixConst := servName + "PathPrefix"

	commentGenerator := types.NewGoComment()
	commentGenerator.Pf("%s is used for all URL paths on a %s server.", pathPrefixConst, servName)
	commentGenerator.Pf("Requests are always: POST %s /method", pathPrefixConst)
	commentGenerator.P("It can be used in an HTTP mux to route requests")

	constGenerator, err := types.NewGoConst(pathPrefixConst, types.String, strconv.Quote(pathPrefix(file, service)), commentGenerator)
	if err != nil {
		return nil, err
	}

	if err := goFile.Const(constGenerator); err != nil {
		return nil, err
	}

	comment := "ServeHTTP implements http.Handler."
	method, err := types.NewGoMethod("s", fmt.Sprintf("*%s", structGenerator.StructMetaData.Name), "ServeHTTP", []*types.Parameter{
		{
			NameOfParameter: "resp",
			Typ:             types.NewUnsafeTypeReference("http.ResponseWriter"),
		},
		{
			NameOfParameter: "req",
			Typ:             types.NewUnsafeTypeReference("*http.Request"),
		},
	}, nil, comment)

	if err != nil {
		return nil, err
	}

	method.DefAssginCall([]string{"ctx"}, types.NewUnsafeTypeReference("req.Context"), nil)
	method.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("xcontext.WithPackageName"), []string{"ctx", `"` + pkgName + `"`})
	method.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("xcontext.WithServiceName"), []string{"ctx", `"` + servName + `"`})
	method.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("xcontext.WithResponseWriter"), []string{"ctx", "resp"})
	method.DefLongVar("err", "error")
	method.DefCall([]string{"ctx", "err"}, types.NewUnsafeTypeReference("transport.CallRequestReceived"), []string{"ctx", "s.hooks"})
	method.DefIfBegin("err", token.NEQ, "nil")
	method.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "err"})
	method.Return()
	method.CloseIf()

	method.DefIfBegin("req.Method", token.NEQ, "http.MethodPost")
	method.DefAssginCall([]string{"msg"}, types.NewUnsafeTypeReference("fmt.Sprintf"), []string{`"unsupported method %q (only POST is allowed)"`, "req.Method"})
	method.DefAssginCall([]string{"terr"}, types.NewUnsafeTypeReference("errors.BadRouteError"), []string{"msg", "req.Method", "req.URL.Path"})
	method.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "terr"})
	method.Return()
	method.CloseIf()

	switchGenerator, err := types.NewSwitchGenerator("req.URL.Path")
	if err != nil {
		return nil, err
	}

	for _, method := range service.Method {
		path := pathFor(file, service, method)
		methName := "serve" + types.CamelCase(method.GetName())
		caseGenerator, err := types.NewCaseGenerator(strconv.Quote(path))
		if err != nil {
			return nil, err
		}

		caseGenerator.Caller(types.NewUnsafeTypeReference(fmt.Sprintf("s.%s", methName)), []string{"ctx", "resp", "req"})
		caseGenerator.Return()
		switchGenerator.Case(*caseGenerator)
	}

	defaultCaseGenerator, err := types.NewDefaultCaseGenerator()
	if err != nil {
		return nil, err
	}
	defaultCaseGenerator.DefAssginCall([]string{"msg"}, types.NewUnsafeTypeReference("fmt.Sprintf"), []string{`"no handler for path %q"`, "req.URL.Path"})
	defaultCaseGenerator.DefAssginCall([]string{"terr"}, types.NewUnsafeTypeReference("errors.BadRouteError"), []string{"msg", "req.Method", "req.URL.Path"})
	defaultCaseGenerator.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "terr"})
	defaultCaseGenerator.Return()

	switchGenerator.Default(*defaultCaseGenerator)
	if err := method.TypeSwitch(*switchGenerator); err != nil {
		return nil, err
	}

	structGenerator.AddMethod(method)
	return structGenerator, nil
}

func (a *API) generateServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto, structGenerator *types.StructGenerator) (*types.StructGenerator, error) {
	methName := types.CamelCase(method.GetName())
	methNameServe := fmt.Sprintf("serve%s", methName)

	comment := fmt.Sprintf("%s is used to set an decoder and encoder for a given content type", methNameServe)
	dispatcherMethod, err := types.NewGoMethod("s", fmt.Sprintf("*%s", structGenerator.StructMetaData.Name), methNameServe, []*types.Parameter{
		{
			NameOfParameter: "ctx",
			Typ:             types.NewUnsafeTypeReference("context.Context"),
		},
		{
			NameOfParameter: "resp",
			Typ:             types.NewUnsafeTypeReference("http.ResponseWriter"),
		},
		{
			NameOfParameter: "req",
			Typ:             types.NewUnsafeTypeReference("*http.Request"),
		},
	}, nil, comment)

	if err != nil {
		return nil, err
	}

	dispatcherMethod.DefAssginCall([]string{"header"}, types.NewUnsafeTypeReference("req.Header.Get"), []string{"xhttp.ContentTypeHeader"})
	dispatcherMethod.DefAssginCall([]string{"i"}, types.NewUnsafeTypeReference("strings.Index"), []string{"header", `";"`})
	dispatcherMethod.DefIfBegin("i", token.EQL, "-1")
	dispatcherMethod.DefCall([]string{"i"}, types.NewUnsafeTypeReference("len"), []string{"header"})
	dispatcherMethod.CloseIf()

	dispatcherMethod.DefAssginCall([]string{"modifiedHeader"}, types.NewUnsafeTypeReference("strings.ToLower"), []string{"header[:i]"})
	dispatcherMethod.DefCall([]string{"modifiedHeader"}, types.NewUnsafeTypeReference("strings.TrimSpace"), []string{"modifiedHeader"})

	dispatcherMethod.DefIfBegin("modifiedHeader", token.EQL, `xhttp.ApplicationJson`)
	dispatcherMethod.Caller(types.NewUnsafeTypeReference(fmt.Sprintf("s.serve%sContent", methName)), []string{"ctx", "resp", "req", "transport.DecodeJSONRequest", "transport.EncodeJSONResponse"})
	dispatcherMethod.Return(nil)
	dispatcherMethod.DefElseIf("modifiedHeader", token.EQL, `xhttp.ApplicationProtobuf`)
	dispatcherMethod.Caller(types.NewUnsafeTypeReference(fmt.Sprintf("s.serve%sContent", methName)), []string{"ctx", "resp", "req", "transport.DecodePROTORequest", "transport.EncodePROTOResponse"})
	dispatcherMethod.Return(nil)
	dispatcherMethod.Else()
	dispatcherMethod.DefAssginCall([]string{"msg"}, types.NewUnsafeTypeReference("fmt.Sprintf"), []string{`"unexpected Content-Type: %q"`, "header"})
	dispatcherMethod.DefAssginCall([]string{"terr"}, types.NewUnsafeTypeReference("errors.BadRouteError"), []string{"msg", "req.Method", "req.URL.Path"})
	dispatcherMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "terr"})
	dispatcherMethod.CloseIf()

	structGenerator.AddMethod(dispatcherMethod)

	structGenerator, err = a.generateServerServeMethod(service, method, structGenerator)
	if err != nil {
		return nil, err
	}

	return structGenerator, nil
}

func (a *API) generateServerServeMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto, structGenerator *types.StructGenerator) (*types.StructGenerator, error) {
	methName := types.CamelCase(method.GetName())
	methServe := fmt.Sprintf("serve%sContent", methName)

	comment := fmt.Sprintf("%s sends object to requester", methServe)
	serveMethod, err := types.NewGoMethod("s", fmt.Sprintf("*%s", structGenerator.StructMetaData.Name), fmt.Sprintf("serve%sContent", methName), []*types.Parameter{
		{
			NameOfParameter: "ctx",
			Typ:             types.NewUnsafeTypeReference("context.Context"),
		},
		{
			NameOfParameter: "resp",
			Typ:             types.NewUnsafeTypeReference("http.ResponseWriter"),
		},
		{
			NameOfParameter: "req",
			Typ:             types.NewUnsafeTypeReference("*http.Request"),
		},
		{
			NameOfParameter: "decodeRequest",
			Typ:             types.NewUnsafeTypeReference("transport.DecodeRequestFunc"),
		},
		{
			NameOfParameter: "encodeResponse",
			Typ:             types.NewUnsafeTypeReference("transport.EncodeResponseFunc"),
		},
	}, nil, comment)

	if err != nil {
		return nil, err
	}

	inputType, err := a.goTypeName(method.GetInputType())
	if err != nil {
		return nil, err
	}

	outputType, err := a.goTypeName(method.GetOutputType())
	if err != nil {
		return nil, err
	}

	// todo wrap call with error catcher

	serveMethod.DefLongVar("err", "error")
	serveMethod.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("xcontext.WithMethodName"), []string{"ctx", `"` + methName + `"`})
	serveMethod.DefCall([]string{"ctx", "err"}, types.NewUnsafeTypeReference("transport.CallRequestRouted"), []string{"ctx", "s.hooks"})
	serveMethod.DefIfBegin("err", token.NEQ, "nil")
	serveMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "err"})
	serveMethod.Return()
	serveMethod.CloseIf()
	serveMethod.Defer(types.NewUnsafeTypeReference("transport.Closebody"), []string{"req.Body", "s.logErrorFunc"})

	serveMethod.DefNew("reqContent", types.NewUnsafeTypeReference(inputType))
	s, _ := serveMethod.SCallWithDefVar([]string{"err"}, types.NewUnsafeTypeReference("decodeRequest"), []string{"ctx", "req", "reqContent"})
	serveMethod.DefIfWithOwnScopeBegin(s, "err", token.NEQ, "nil")
	serveMethod.Caller(types.NewUnsafeTypeReference("s.logErrorFunc"), []string{`"%v"`, "err"})
	serveMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "err"})
	serveMethod.Return()
	serveMethod.CloseIf()

	responseCallWrapper, _ := types.NewAnonymousGoFunc("endpointWrapper", nil, []types.TypeReference{types.NewUnsafeTypeReference(fmt.Sprintf("*%s", outputType)), types.NewUnsafeTypeReference("error")})
	responseDeferWrapper, _ := types.NewAnonymousGoFunc("deferWrapper", nil, nil)

	s, _ = responseDeferWrapper.SCallWithDefVar([]string{"r"}, types.NewUnsafeTypeReference("recover"), nil)
	responseDeferWrapper.DefIfWithOwnScopeBegin(s, "r", token.NEQ, "nil")
	responseDeferWrapper.DefAssginCall([]string{"terr"}, types.NewUnsafeTypeReference("errors.InternalError"), []string{`"Internal service panic"`})
	responseDeferWrapper.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "terr"})
	responseDeferWrapper.Caller(types.NewUnsafeTypeReference("panic"), []string{"r"})
	responseDeferWrapper.CloseIf()
	responseCallWrapper.AnonymousGoFunc(responseDeferWrapper)
	responseCallWrapper.Defer(types.NewUnsafeTypeReference("deferWrapper"), nil)
	responseCallWrapper.ReturnCaller(types.NewUnsafeTypeReference(fmt.Sprintf("s.%s", methName)), []string{"ctx", "reqContent"})
	serveMethod.AnonymousGoFunc(responseCallWrapper)
	serveMethod.DefAssginCall([]string{"respContent", "err"}, types.NewUnsafeTypeReference("endpointWrapper"), nil)
	serveMethod.DefIfBegin("err", token.NEQ, "nil")
	serveMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "err"})
	serveMethod.Return()
	serveMethod.CloseIf()

	serveMethod.DefIfBegin("respContent", token.EQL, "nil")
	msg := fmt.Sprintf(`"received a nil * %s, and nil error while calling %s. nil responses are not supported"`, outputType, methName)
	serveMethod.DefAssginCall([]string{"terr"}, types.NewUnsafeTypeReference("errors.InternalError"), []string{msg})
	serveMethod.Caller(types.NewUnsafeTypeReference("s.logErrorFunc"), []string{`"%v"`, "err"})
	serveMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "terr"})
	serveMethod.Return()
	serveMethod.CloseIf()

	serveMethod.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("transport.CallResponsePrepared"), []string{"ctx", "s.hooks"})
	s, _ = serveMethod.SCallWithDefVar([]string{"err"}, types.NewUnsafeTypeReference("encodeResponse"), []string{"ctx", "resp", "respContent"})
	serveMethod.DefIfWithOwnScopeBegin(s, "err", token.NEQ, "nil")
	serveMethod.Caller(types.NewUnsafeTypeReference("s.logErrorFunc"), []string{`"%v"`, "err"})
	serveMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "err"})
	serveMethod.Return()
	serveMethod.CloseIf()

	serveMethod.Caller(types.NewUnsafeTypeReference("transport.CallResponseSent"), []string{"ctx", "s.hooks"})

	structGenerator.AddMethod(serveMethod)

	return structGenerator, nil
}

func (a *API) generateServiceMetadataAccessors(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, structGenerator *types.StructGenerator) (*types.StructGenerator, error) {
	index := 0
	for i, s := range file.Service {
		if s.GetName() == service.GetName() {
			index = i
		}
	}

	structName := structGenerator.StructMetaData.Name

	comment := "ServiceDescriptor describes an service."
	serviceDescriptorMethod, err := types.NewGoMethod("s", fmt.Sprintf("*%s", structName), "ServiceDescriptor", nil,
		[]types.TypeReference{
			types.TypeReferenceFromInstance([]byte(nil)),
			types.TypeReferenceFromInstance(int(0)),
		}, comment)

	if err != nil {
		return nil, err
	}

	serviceDescriptorMethod.Return([]string{a.serviceMetadataVarName(), strconv.Itoa(index)})
	structGenerator.AddMethod(serviceDescriptorMethod)

	comment = "ProtocGenXServiceVersion returns which xservice version was used to generate that service"
	protocGenXServiceVersionMethod, err := types.NewGoMethod("s", fmt.Sprintf("*%s", structName), "ProtocGenXServiceVersion", nil,
		[]types.TypeReference{
			types.TypeReferenceFromInstance(string("")),
		}, comment)

	if err != nil {
		return nil, err
	}

	protocGenXServiceVersionMethod.Return([]string{strconv.Quote(xproto.Version)})
	structGenerator.AddMethod(protocGenXServiceVersionMethod)

	return structGenerator, nil
}

func (a *API) generateFileDescriptor(file *descriptor.FileDescriptorProto, goFile *types.FileGenerator) (*types.FileGenerator, error) {
	// Copied straight of of protoc-gen-go, which trims out comments.
	pb := proto.Clone(file).(*descriptor.FileDescriptorProto)
	pb.SourceCodeInfo = nil

	b, err := proto.Marshal(pb)
	if err != nil {
		return nil, err
	}

	var descriptorProto bytes.Buffer
	w, err := gzip.NewWriterLevel(&descriptorProto, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	w.Write(b)
	w.Close()
	b = descriptorProto.Bytes()

	v := a.serviceMetadataVarName()

	buff := new(bytes.Buffer)

	buff.WriteString(fmt.Sprintf("// %d bytes of a gzipped FileDescriptorProto \n", len(b)))
	buff.WriteString(fmt.Sprintf("var %s = []byte{", v))
	for len(b) > 0 {
		n := 16
		if n > len(b) {
			n = len(b)
		}

		s := ""
		for _, c := range b[:n] {
			s += fmt.Sprintf("0x%02x,", c)
		}
		buff.WriteString(s)

		b = b[n:]
	}
	buff.WriteString("}")

	goFile.Var(buff.String())

	return goFile, nil
}

// pathPrefix returns the base path for all methods handled by a particular
// service. It includes a trailing slash. (for example
// "/xservice/example.Haberdasher/").
func pathPrefix(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	return fmt.Sprintf("/xservice/%s/", fullServiceName(file, service))
}

// pathFor returns the complete path for requests to a particular method on a
// particular service.
func pathFor(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) string {
	return pathPrefix(file, service) + types.CamelCase(method.GetName())
}

// Given a protobuf name for a Message, return the Go name we will use for that
// type, including its package prefix.
func (a *API) goTypeName(protoName string) (string, error) {
	def := a.reg.MessageDefinition(protoName)
	if def == nil {
		return "", errors.Errorf("could not find message for %s", protoName)
	}

	var prefix string
	if pkg := a.goPackageName(def.File); pkg != a.genPkgName {
		prefix = pkg + "."
	}

	var name string
	for _, parent := range def.Lineage() {
		name += parent.Descriptor.GetName() + "_"
	}
	name += def.Descriptor.GetName()
	return prefix + name, nil
}

func (a *API) goPackageName(file *descriptor.FileDescriptorProto) string {
	return a.fileToGoPackageName[file]
}

func unexported(s string) string { return strings.ToLower(s[:1]) + s[1:] }

func fullServiceName(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	name := types.CamelCase(service.GetName())
	if pkg := pkgName(file); pkg != "" {
		name = pkg + "." + name
	}
	return name
}

func pkgName(file *descriptor.FileDescriptorProto) string {
	return file.GetPackage()
}

func serviceName(service *descriptor.ServiceDescriptorProto) string {
	return types.CamelCase(service.GetName())
}

func serviceStruct(service *descriptor.ServiceDescriptorProto) string {
	return unexported(serviceName(service)) + "Server"
}

func methodName(method *descriptor.MethodDescriptorProto) string {
	return types.CamelCase(method.GetName())
}

func fileDescSliceContains(slice []*descriptor.FileDescriptorProto, f *descriptor.FileDescriptorProto) bool {
	for _, sf := range slice {
		if f == sf {
			return true
		}
	}
	return false
}

var EmptyComment error = errors.New("comment is empty")

func prepareComment(comments typemap.DefinitionComments) (string, error) {
	text := strings.TrimSuffix(comments.Leading, "\n")
	if len(strings.TrimSpace(text)) == 0 {
		return "", EmptyComment
	}
	comment := types.NewGoComment()
	split := strings.Split(text, "\n")
	for _, line := range split {
		comment.P(strings.TrimPrefix(line, " "))
	}

	commentRendered, err := comment.Render()
	if err != nil {
		return "", err
	}

	return commentRendered, nil
}

// serviceMetadataVarName is the variable name used in generated code to refer
// to the compressed bytes of this descriptor. It is not exported, so it is only
// valid inside the generated package.
//
// protoc-gen-go writes its own version of this file, but so does
// protoc-gen-gogo - with a different name! Twirp aims to be compatible with
// both; the simplest way forward is to write the file descriptor again as
// another variable that we control.
func (a *API) serviceMetadataVarName() string {
	return fmt.Sprintf("xserviceFileDescriptor%d", a.filesHandled)
}

// deduceGenPkgName figures out the go package name to use for generated code.
// Will try to use the explicit go_package setting in a file (if set, must be
// consistent in all files). If no files have go_package set, then use the
// protobuf package name (must be consistent in all files)
func deduceGenPkgName(genFiles []*descriptor.FileDescriptorProto) (string, error) {
	var genPkgName string
	for _, f := range genFiles {
		name, explicit := xprotoutil.GoPackageName(f)
		name = types.BaseName(name)
		if explicit {
			name = types.Identifier(name)
			if genPkgName != "" && genPkgName != name {
				// Make sure they're all set consistently.
				return "", errors.Errorf("files have conflicting go_package settings, must be the same: %q and %q", genPkgName, name)
			}
			genPkgName = name
		}
	}
	if genPkgName != "" {
		return genPkgName, nil
	}

	// If there is no explicit setting, then check the implicit package name
	// (derived from the protobuf package name) of the files and make sure it's
	// consistent.
	for _, f := range genFiles {
		name, _ := xprotoutil.GoPackageName(f)
		name = types.BaseName(name)
		name = types.Identifier(name)
		if genPkgName != "" && genPkgName != name {
			return "", errors.Errorf("files have conflicting package names, must be the same or overridden with go_package: %q and %q", genPkgName, name)
		}
		genPkgName = name
	}

	// All the files have the same name, so we're good.
	return genPkgName, nil
}
