package goproto

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/donutloop/xservice/internal/xgenerator/types"
	"github.com/donutloop/xservice/internal/xproto"
	"github.com/donutloop/xservice/internal/xproto/typesmap"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pkg/errors"
	"go/token"
	"strconv"
	"strings"
)

type Server struct {
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

func NewServerGenerator() *Server {
	gen := &Server{
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

func (t *Server) Generate(in *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	t.genFiles = FilesToGenerate(in)

	// Collect information on types.
	t.reg = typemap.New(in.ProtoFile)

	// Register names of packages that we import.

	// Time to figure out package names of objects defined in protobuf. First,
	// we'll figure out the name for the package we're generating.
	genPkgName, err := deduceGenPkgName(t.genFiles)
	if err != nil {
		return nil, errors.Wrap(err, "todo")
	}
	t.genPkgName = genPkgName

	// Next, we need to pick names for all the files that are dependencies.
	for _, f := range in.ProtoFile {
		if fileDescSliceContains(t.genFiles, f) {
			// This is a file we are generating. It gets the shared package name.
			t.fileToGoPackageName[f] = t.genPkgName
		} else {
			// This is a dependency. Use its package name.
			name := f.GetPackage()
			if name == "" {
				name = types.BaseName(f.GetName())
			}
			name = types.Identifier(name)
			t.fileToGoPackageName[f] = name
		}
	}

	// Showtime! Generate the response.
	resp := new(plugin.CodeGeneratorResponse)
	for _, f := range t.genFiles {
		respFile, err := t.generate(f)
		if err != nil {
			return nil, err
		}
		if respFile != nil {
			resp.File = append(resp.File, respFile)
		}
	}
	return resp, nil
}

func (s *Server) generate(fileDescriptor *descriptor.FileDescriptorProto) (*plugin.CodeGeneratorResponse_File, error) {
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

	s.generateFileHeader(fileDescriptor, goFile)

	s.generateImports(fileDescriptor, goFile)

	// For each service, generate client stubs and server
	for i, service := range fileDescriptor.Service {
		goFile, err = s.generateService(fileDescriptor, service, goFile, i)
		if err != nil {
			return nil, err
		}
	}

	goFile, err = s.generateFileDescriptor(fileDescriptor, goFile)
	if err != nil {
		return nil, err
	}

	resp.Name = proto.String(goFile.GetFileName())

	content, err := goFile.RenderAndFormatCode()
	if err != nil {
		return nil, err
	}

	resp.Content = proto.String(string(content))

	s.filesHandled++
	return resp, nil
}

// deduceGenPkgName figures out the go package name to use for generated code.
// Will try to use the explicit go_package setting in a file (if set, must be
// consistent in all files). If no files have go_package set, then use the
// protobuf package name (must be consistent in all files)
func deduceGenPkgName(genFiles []*descriptor.FileDescriptorProto) (string, error) {
	var genPkgName string
	for _, f := range genFiles {
		name, explicit := goPackageName(f)
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
		name, _ := goPackageName(f)
		name = types.Identifier(name)
		if genPkgName != "" && genPkgName != name {
			return "", errors.Errorf("files have conflicting package names, must be the same or overridden with go_package: %q and %q", genPkgName, name)
		}
		genPkgName = name
	}

	// All the files have the same name, so we're good.
	return genPkgName, nil
}

// goPackageName returns the Go package name to use in the generated Go file.
// The result explicitly reports whether the name came from an option go_package
// statement. If explicit is false, the name was derived from the protocol
// buffer's package statement or the input file name.
func goPackageName(f *descriptor.FileDescriptorProto) (name string, explicit bool) {
	// Does the file have a "go_package" option?
	if _, pkg, ok := goPackageOption(f); ok {
		return pkg, true
	}

	// Does the file have a package clause?
	if pkg := f.GetPackage(); pkg != "" {
		return pkg, false
	}
	// Use the file base name.
	return types.BaseName(f.GetName()), false
}

// goPackageOption interprets the file's go_package option.
// If there is no go_package, it returns ("", "", false).
// If there's a simple name, it returns ("", pkg, true).
// If the option implies an import path, it returns (impPath, pkg, true).
func goPackageOption(f *descriptor.FileDescriptorProto) (impPath, pkg string, ok bool) {
	pkg = f.GetOptions().GetGoPackage()
	if pkg == "" {
		return
	}
	ok = true
	// The presence of a slash implies there's an import path.
	slash := strings.LastIndex(pkg, "/")
	if slash < 0 {
		return
	}
	impPath, pkg = pkg, pkg[slash+1:]

	// A semicolon-delimited suffix overrides the package name.
	sc := strings.IndexByte(impPath, ';')
	if sc < 0 {
		return
	}
	impPath, pkg = impPath[:sc], impPath[sc+1:]
	return
}

func (s *Server) generateFileHeader(file *descriptor.FileDescriptorProto, goFile *types.FileGenerator) (*types.FileGenerator, error) {

	c := types.NewGoComment()

	c.Pf("// Code generated by xproto %s, DO NOT EDIT.", xproto.Version)
	c.Pf("// source: %s ", file.GetName())
	if s.filesHandled == 0 {
		c.Pf("Package %s is a generated stub package.", s.genPkgName)
		c.Pf("This code was generated with github.com/donutloop/xservice %s", xproto.Version)
		comment, err := s.reg.FileComments(file)
		if err == nil && comment.Leading != "" {
			for _, line := range strings.Split(comment.Leading, "\n") {
				line = strings.TrimPrefix(line, " ")
				// ensure we don't escape from the block comment
				line = strings.Replace(line, "*/", "* /", -1)
				c.P(line)
			}
		}
		c.P("It is generated from these files:")
		for _, f := range s.genFiles {
			c.Pf("\t %s", f.GetName())
		}
	}
	c.Pf(`package %s`, s.genPkgName)

	if err := goFile.HeaderComment(c); err != nil {
		return nil, err
	}

	return goFile, nil
}

func (t *Server) generateImports(file *descriptor.FileDescriptorProto, goFile *types.FileGenerator) {
	if len(file.Service) == 0 {
		return
	}

	//goFile.Import("proto", "github.com/golang/protobuf/proto")
	goFile.Import("jsonpb", "github.com/golang/protobuf/jsonpb")
	goFile.Import("", "github.com/donutloop/xservice/framework/transport")
	goFile.Import("", "github.com/donutloop/xservice/framework/ctxsetters")
	goFile.Import("", "github.com/donutloop/xservice/framework/errors")
	goFile.Import("", "github.com/donutloop/xservice/framework/hooks")
	goFile.Import("", "github.com/donutloop/xservice/framework/server")
	goFile.Import("", "github.com/donutloop/xservice/framework/xhttp")

}

func (s *Server) generateService(fileDescriptor *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, goFile *types.FileGenerator, index int) (*types.FileGenerator, error) {

	var err error

	// interface
	goFile, err = s.generatServiceInterface(fileDescriptor, service, goFile)
	if err != nil {
		return nil, err
	}

	// Server
	goFile, err = s.generateServer(fileDescriptor, service, goFile)
	if err != nil {
		return nil, err
	}

	return goFile, nil
}

func (s *Server) generatServiceInterface(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, goFile *types.FileGenerator) (*types.FileGenerator, error) {

	serviceInterface, err := types.NewGoInterface(serviceName(service))
	if err != nil {
		return nil, err
	}

	comments, err := s.reg.ServiceComments(file, service)
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
		comments, err = s.reg.MethodComments(file, service, method)
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

		inputType, err := s.goTypeName(method.GetInputType())
		if err != nil {
			return nil, err
		}

		outputType, err := s.goTypeName(method.GetOutputType())
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

func (s *Server) generateServer(fileDescriptor *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, goFile *types.FileGenerator) (*types.FileGenerator, error) {
	// Server implementation.
	structGenerator, err := types.NewGoStruct(serviceStruct(service), true)
	if err != nil {
		return nil, err
	}

	structGenerator.Type(types.NewUnsafeTypeReference(serviceName(service)), "")
	structGenerator.AddUnexportedField("hooks", types.NewUnsafeTypeReference("*hooks.ServerHooks"), "")

	goFile, err = s.generateServerConstructor(serviceName(service), structGenerator.StructMetaData.Name, goFile)
	if err != nil {
		return nil, err
	}

	structGenerator, err = s.generateServerWriteError(structGenerator)
	if err != nil {
		return nil, err
	}

	structGenerator, err = s.generateServerRouting(fileDescriptor, service, structGenerator, goFile)
	if err != nil {
		return nil, err
	}

	// Methods.
	for _, method := range service.Method {
		structGenerator, err = s.generateServerMethod(service, method, structGenerator)
		if err != nil {
			return nil, err
		}
	}

	structGenerator, err = s.generateServiceMetadataAccessors(fileDescriptor, service, structGenerator)
	if err != nil {
		return nil, err
	}

	if err := goFile.TypesWithMethods(structGenerator); err != nil {
		return nil, err
	}

	return goFile, nil
}

func (s *Server) generateServerConstructor(serverName, serverStructName string, goFile *types.FileGenerator) (*types.FileGenerator, error) {

	constructorName := fmt.Sprintf("New%sServer", serverName)

	f, err  := types.NewGoFunc(constructorName, []*types.Parameter{
		{
			NameOfParameter: "svc",
			Typ: types.NewUnsafeTypeReference(serverName),
		},
		{
			NameOfParameter: "hooks",
			Typ: types.NewUnsafeTypeReference("*hooks.ServerHooks"),
		},
	},
	[]types.TypeReference{
		types.NewUnsafeTypeReference("server.Server"),
	})
	if err != nil {
		return nil, err
	}

	initStructGenerator, err := types.NewInitGoStruct(serverStructName)
	if err != nil {
		return nil, err
	}

	initStructGenerator.AddExportedValueToField(serverName, "svc")
	initStructGenerator.AddUnexportedValueToField("hooks", "hooks")

	if err := f.InitStruct("return", initStructGenerator, true); err != nil {
		return nil, err
	}

	if err := goFile.Func(f); err != nil {
		return nil, err
	}

	return goFile, nil
}

func (s *Server) generateServerWriteError(structGenerator *types.StructGenerator) (*types.StructGenerator, error) {
	method, err := types.NewGoMethod("s",  fmt.Sprintf("*%s", structGenerator.StructMetaData.Name), "writeError", []*types.Parameter{
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

func (s *Server) generateServerRouting(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, structGenerator *types.StructGenerator, goFile *types.FileGenerator) (*types.StructGenerator, error) {

	pkgName := pkgName(file)
	servName := serviceName(service)
	pathPrefixConst := servName + "PathPrefix"

	commentGenerator := types.NewGoComment()
	commentGenerator.Pf("%s is used for all URL paths on a %s server.", pathPrefixConst, servName)
	commentGenerator.Pf("Requests are always: POST %s /method",  pathPrefixConst)
	commentGenerator.P("It can be used in an HTTP mux to route requests")

	constGenerator, err := types.NewGoConst(pathPrefixConst, types.String, strconv.Quote(pathPrefix(file, service)), commentGenerator)
	if err != nil {
		return nil, err
	}

	goFile.Const(constGenerator)


	method, err := types.NewGoMethod("s",  fmt.Sprintf("*%s", structGenerator.StructMetaData.Name), "ServeHTTP", []*types.Parameter{
		{
			NameOfParameter: "resp",
			Typ:             types.NewUnsafeTypeReference("http.ResponseWriter"),
		},
		{
			NameOfParameter: "req",
			Typ:             types.NewUnsafeTypeReference("*http.Request"),
		},
	}, nil, "")

	if err != nil {
		return nil, err
	}


	method.DefAssginCall([]string{"ctx"}, types.NewUnsafeTypeReference("req.Context"), nil)
	method.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("ctxsetters.WithPackageName"), []string{"ctx", `"` + pkgName + `"`})
	method.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("ctxsetters.WithServiceName"), []string{"ctx", `"` + servName + `"`})
	method.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("ctxsetters.WithResponseWriter"), []string{"ctx", "resp"})
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

func (s *Server) generateServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto, structGenerator *types.StructGenerator) (*types.StructGenerator, error) {
	methName := types.CamelCase(method.GetName())

	dispatcherMethod, err := types.NewGoMethod("s", fmt.Sprintf("*%s", structGenerator.StructMetaData.Name), fmt.Sprintf("serve%s", methName), []*types.Parameter{
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
	}, nil, "")

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

	dispatcherMethod.DefIfBegin("modifiedHeader", token.EQL, `"application/json"`)
	dispatcherMethod.Caller(types.NewUnsafeTypeReference(fmt.Sprintf("s.serve%sJSON", methName)), []string{"ctx", "resp", "req"})
	dispatcherMethod.Else()
	dispatcherMethod.DefAssginCall([]string{"msg"}, types.NewUnsafeTypeReference("fmt.Sprintf"), []string{`"unexpected Content-Type: %q"`, "header"})
	dispatcherMethod.DefAssginCall([]string{"terr"}, types.NewUnsafeTypeReference("errors.BadRouteError"), []string{"msg", "req.Method", "req.URL.Path"})
	dispatcherMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "terr"})
	dispatcherMethod.CloseIf()

	dispatcherMethod.Return(nil)
	structGenerator.AddMethod(dispatcherMethod)

	structGenerator, err = s.generateServerJSONMethod(service, method, structGenerator)
	if err != nil {
		return nil, err
	}
	return structGenerator, nil
}

func (s *Server) generateServerJSONMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto, structGenerator *types.StructGenerator) (*types.StructGenerator, error) {
	methName := types.CamelCase(method.GetName())

	serveMethod, err := types.NewGoMethod("s",  fmt.Sprintf("*%s", structGenerator.StructMetaData.Name), fmt.Sprintf("serve%sJSON", methName), []*types.Parameter{
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
	}, nil, "")

	if err != nil {
		return nil, err
	}

	inputType, err := s.goTypeName(method.GetInputType())
	if err != nil {
		return nil, err
	}

	outputType, err := s.goTypeName(method.GetOutputType())
	if err != nil {
		return nil, err
	}

	// todo wrap call with error catcher

	serveMethod.DefLongVar("err", "error")
	serveMethod.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("ctxsetters.WithMethodName"), []string{"ctx", `"` + methName + `"`})
	serveMethod.DefCall([]string{"ctx", "err"}, types.NewUnsafeTypeReference("transport.CallRequestRouted"), []string{"ctx", "s.hooks"})
	serveMethod.DefIfBegin("err", token.NEQ, "nil")
	serveMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "err"})
	serveMethod.Return()
	serveMethod.CloseIf()
	serveMethod.Defer(types.NewUnsafeTypeReference("transport.Closebody"), []string{"req.Body"})
	serveMethod.DefNew("reqContent", types.NewUnsafeTypeReference(inputType))
	serveMethod.DefShortVar("unmarshaler", "jsonpb.Unmarshaler{AllowUnknownFields: true}")
	serveMethod.DefCall([]string{"err"}, types.NewUnsafeTypeReference("unmarshaler.Unmarshal"), []string{"req.Body", "reqContent"})
	serveMethod.DefIfBegin("err", token.NEQ, "nil")
	serveMethod.DefCall([]string{"err"}, types.NewUnsafeTypeReference("errors.WrapErr"), []string{"err", `"failed to parse request json"`})
	serveMethod.DefAssginCall([]string{"terr"}, types.NewUnsafeTypeReference("errors.InternalErrorWith"), []string{"err"})
	serveMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "terr"})
	serveMethod.Return()
	serveMethod.CloseIf()
	serveMethod.DefNew("respContent", types.NewUnsafeTypeReference(outputType))
	responseCallWrapper, _ := types.NewAnonymousGoFunc("responseCallWrapper", nil, nil)
	responseDeferWrapper, _ := types.NewAnonymousGoFunc("responseDeferWrapper", nil, nil)

	responseDeferWrapper.DefAssginCall([]string{"r"}, types.NewUnsafeTypeReference("recover"), nil)
	responseDeferWrapper.DefIfBegin("r", token.NEQ, "nil")
	responseDeferWrapper.DefAssginCall([]string{"terr"}, types.NewUnsafeTypeReference("errors.InternalError"), []string{`"Internal serivce panic"`})
	responseDeferWrapper.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "terr"})
	responseDeferWrapper.CloseIf()
	responseDeferWrapper.Caller(types.NewUnsafeTypeReference("panic"), []string{"r"})
	responseCallWrapper.AnonymousGoFunc(responseDeferWrapper)
	responseCallWrapper.Defer(types.NewUnsafeTypeReference("responseDeferWrapper"), nil)
	responseCallWrapper.DefCall([]string{"respContent", "err"}, types.NewUnsafeTypeReference(fmt.Sprintf("s.%s", methName)), []string{"ctx", "reqContent"})
	serveMethod.AnonymousGoFunc(responseCallWrapper)
	serveMethod.Defer(types.NewUnsafeTypeReference("responseCallWrapper"), nil)
	serveMethod.DefIfBegin("err", token.NEQ, "nil")
	serveMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "err"})
	serveMethod.Return()
	serveMethod.CloseIf()

	serveMethod.DefIfBegin("respContent", token.EQL, "nil")
	msg := fmt.Sprintf(`"received a nil * %s, and nil error while calling %s. nil responses are not supported"`, outputType, methName)
	serveMethod.DefAssginCall([]string{"terr"}, types.NewUnsafeTypeReference("errors.InternalError"), []string{msg})
	serveMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "terr"})
	serveMethod.Return()
	serveMethod.CloseIf()

	serveMethod.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("transport.CallResponsePrepared"), []string{"ctx", "s.hooks"})
	serveMethod.DefNew("buff", types.NewUnsafeTypeReference("bytes.Buffer"))
	serveMethod.DefShortVar("marshaler", "&jsonpb.Marshaler{OrigName: true}")
	serveMethod.DefCall([]string{"err"}, types.NewUnsafeTypeReference("marshaler.Marshal"), []string{"buff", "respContent"})
	serveMethod.DefIfBegin("err", token.NEQ, "nil")
	serveMethod.DefCall([]string{"err"}, types.NewUnsafeTypeReference("errors.WrapErr"), []string{"err", `"failed to marshal json response"`})
	serveMethod.DefAssginCall([]string{"terr"}, types.NewUnsafeTypeReference("errors.InternalErrorWith"), []string{"err"})
	serveMethod.Caller(types.NewUnsafeTypeReference("s.writeError"), []string{"ctx", "resp", "terr"})
	serveMethod.Return()
	serveMethod.CloseIf()

	serveMethod.DefCall([]string{"ctx"}, types.NewUnsafeTypeReference("ctxsetters.WithStatusCode"), []string{"ctx", "http.StatusOK"})
	serveMethod.Caller(types.NewUnsafeTypeReference("req.Header.Set"), []string{"xhttp.ContentTypeHeader", "xhttp.ApplicationJson"})
	serveMethod.Caller(types.NewUnsafeTypeReference("resp.WriteHeader"), []string{"http.StatusOK"})
	serveMethod.DefAssginCall([]string{"respBytes"}, types.NewUnsafeTypeReference("buff.Bytes"), nil)

	serveMethod.DefCall([]string{"_", "err"}, types.NewUnsafeTypeReference("resp.Write"), []string{"respBytes"})

	serveMethod.DefIfBegin("err", token.NEQ, "nil")
	// t.P(`    `, t.pkgs["log"], `.Printf("errored while writing response to client, but already sent response status code to 200: %s", err)`)
	serveMethod.Return()
	serveMethod.CloseIf()

	serveMethod.Caller(types.NewUnsafeTypeReference("transport.CallResponseSent"), []string{"ctx", "s.hooks"})

	structGenerator.AddMethod(serveMethod)

	return structGenerator, nil
}

func (t *Server) generateServiceMetadataAccessors(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, structGenerator *types.StructGenerator) (*types.StructGenerator, error) {
	index := 0
	for i, s := range file.Service {
		if s.GetName() == service.GetName() {
			index = i
		}
	}

	structName := structGenerator.StructMetaData.Name

	serviceDescriptorMethod, err := types.NewGoMethod("s",  fmt.Sprintf("*%s", structName), "ServiceDescriptor", nil,
		[]types.TypeReference{
			types.TypeReferenceFromInstance([]byte(nil)),
			types.TypeReferenceFromInstance(int(0)),
		}, "")

	if err != nil {
		return nil, err
	}

	serviceDescriptorMethod.Return([]string{t.serviceMetadataVarName(), strconv.Itoa(index)})
	structGenerator.AddMethod(serviceDescriptorMethod)

	protocGenXServiceVersionMethod, err := types.NewGoMethod("s",  fmt.Sprintf("*%s", structName), "ProtocGenXServiceVersion", nil,
		[]types.TypeReference{
			types.TypeReferenceFromInstance(string("")),
		}, "")

	if err != nil {
		return nil, err
	}

	protocGenXServiceVersionMethod.Return([]string{strconv.Quote(xproto.Version)})
	structGenerator.AddMethod(protocGenXServiceVersionMethod)

	return structGenerator, nil
}

func (t *Server) generateFileDescriptor(file *descriptor.FileDescriptorProto, goFile *types.FileGenerator) (*types.FileGenerator, error) {
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

	v := t.serviceMetadataVarName()

	buff := new(bytes.Buffer)

	buff.WriteString(fmt.Sprintf("// %d bytes of a gzipped FileDescriptorProto \n", len(b)))
	buff.WriteString(fmt.Sprintf("var %s = []byte{",  v))
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
func (s *Server) goTypeName(protoName string) (string, error) {
	def := s.reg.MessageDefinition(protoName)
	if def == nil {
		return "", errors.Errorf("could not find message for %s", protoName)
	}

	var prefix string
	if pkg := s.goPackageName(def.File); pkg != s.genPkgName {
		prefix = pkg + "."
	}

	var name string
	for _, parent := range def.Lineage() {
		name += parent.Descriptor.GetName() + "_"
	}
	name += def.Descriptor.GetName()
	return prefix + name, nil
}

func (t *Server) goPackageName(file *descriptor.FileDescriptorProto) string {
	return t.fileToGoPackageName[file]
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
func (t *Server) serviceMetadataVarName() string {
	return fmt.Sprintf("xserviceFileDescriptor%d", t.filesHandled)
}
