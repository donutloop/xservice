package proto

import (
	"github.com/donutloop/xservice/internal/xproto/typesmap"
	"bytes"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"fmt"
	"compress/gzip"
	"strconv"
	"path"
	"github.com/donutloop/xservice/internal/xgenerator/types"
	"github.com/pkg/errors"
	"strings"
	"github.com/gogo/protobuf/proto"
	"github.com/donutloop/xservice/internal/xproto"
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

func newServerGenerator() *Server {
	gen := &Server{
		pkgs:                make(map[string]string),
		pkgNamesInUse:       make(map[string]bool),
		fileToGoPackageName: make(map[*descriptor.FileDescriptorProto]string),
		output:              bytes.NewBuffer(nil),
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

	goFile, err := types.NewGoFile(*fileDescriptor.Package, types.GoFileName(*fileDescriptor.Package), "")
	if err != nil {
		return nil, nil
	}

	s.generateFileHeader(file)

	s.generateImports(file)

	// For each service, generate client stubs and server
	for i, service := range file.Service {
		s.generateService(file, service, i)
	}

	// Util functions only generated once per package
	if s.filesHandled == 0 {
		s.generateUtils()
	}

	s.generateFileDescriptor(file)



	resp.Name = proto.String(goFileName(file))
	resp.Content = proto.String(s.formattedOutput())
	s.output.Reset()

	s.filesHandled++
	return resp
}

// deduceGenPkgName figures out the go package name to use for generated code.
// Will try to use the explicit go_package setting in a file (if set, must be
// consistent in all files). If no files have go_package set, then use the
// protobuf package name (must be consistent in all files)
func  deduceGenPkgName(genFiles []*descriptor.FileDescriptorProto) (string, error) {
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


func (s *Server) generateFileHeader(file *descriptor.FileDescriptorProto, goFile *types.FileGenerator) {

	goFile.HeaderCommentF("// Code generated by xproto %s, DO NOT EDIT.", xproto.Version)
	goFile.HeaderCommentF("// source: %s ", file.GetName())
	goFile.HeaderComment("\n")
	if s.filesHandled == 0 {
		goFile.HeaderComment("/*")
		t.P("Package ", s.genPkgName, " is a generated twirp stub package.")
		t.P("This code was generated with github.com/donutloop/xservice ", xproto.Version, ".")
		t.P()
		comment, err := t.reg.FileComments(file)
		if err == nil && comment.Leading != "" {
			for _, line := range strings.Split(comment.Leading, "\n") {
				line = strings.TrimPrefix(line, " ")
				// ensure we don't escape from the block comment
				line = strings.Replace(line, "*/", "* /", -1)
				t.P(line)
			}
			t.P()
		}
		t.P("It is generated from these files:")
		for _, f := range t.genFiles {
			t.P("\t", f.GetName())
		}
		t.P("*/")
	}
	t.P(`package `, t.genPkgName)
	t.P()
}

func (t *Server) generateImports(file *descriptor.FileDescriptorProto) {
	if len(file.Service) == 0 {
		return
	}
	t.P(`import `, t.pkgs["jsonpb"], ` "github.com/golang/protobuf/jsonpb"`)
	t.P(`import `, t.pkgs["proto"], ` "github.com/golang/protobuf/proto"`)

	// It's legal to import a message and use it as an input or output for a
	// method. Make sure to import the package of any such message. First, dedupe
	// them.
	deps := make(map[string]string) // Map of package name to quoted import path.
	ourImportPath := path.Dir(goFileName(file))
	for _, s := range file.Service {
		for _, m := range s.Method {
			defs := []*typemap.MessageDefinition{
				t.reg.MethodInputDefinition(m),
				t.reg.MethodOutputDefinition(m),
			}
			for _, def := range defs {
				importPath := path.Dir(goFileName(def.File))
				if importPath != ourImportPath {
					pkg := t.goPackageName(def.File)
					deps[pkg] = strconv.Quote(importPath)
				}
			}
		}
	}
	for pkg, importPath := range deps {
		t.P(`import `, pkg, ` `, importPath)
	}
	if len(deps) > 0 {
		t.P()
	}
}

// Generate utility functions used in Twirp code.
// These should be generated just once per package.
func (t *Server) generateUtils() {

}

func (t *Server) generateService(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, index int) {
	servName := serviceName(service)

	t.sectionComment(servName + ` Interface`)
	t.generateTwirpInterface(file, service)

	t.sectionComment(servName + ` Protobuf Client`)
	t.generateClient("Protobuf", file, service)

	t.sectionComment(servName + ` JSON Client`)
	t.generateClient("JSON", file, service)

	// Server
	t.sectionComment(servName + ` Server Handler`)
	t.generateServer(file, service)
}

func (t *Server) generateTwirpInterface(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	servName := serviceName(service)

	comments, err := t.reg.ServiceComments(file, service)
	if err == nil {
		t.printComments(comments)
	}
	t.P(`type `, servName, ` interface {`)
	for _, method := range service.Method {
		comments, err = t.reg.MethodComments(file, service, method)
		if err == nil {
			t.printComments(comments)
		}
		t.P(t.generateSignature(method))
		t.P()
	}
	t.P(`}`)
}

func (t *server) generateSignature(method *descriptor.MethodDescriptorProto) string {
	methName := methodName(method)
	inputType := t.goTypeName(method.GetInputType())
	outputType := t.goTypeName(method.GetOutputType())
	return fmt.Sprintf(`	%s(%s.Context, *%s) (*%s, error)`, methName, t.pkgs["context"], inputType, outputType)
}


func (t *server) generateServer(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	servName := serviceName(service)

	// Server implementation.
	servStruct := serviceStruct(service)

	// Routing.
	t.generateServerRouting(servStruct, file, service)

	// Methods.
	for _, method := range service.Method {
		t.generateServerMethod(service, method)
	}

	t.generateServiceMetadataAccessors(file, service)
}

// pathPrefix returns the base path for all methods handled by a particular
// service. It includes a trailing slash. (for example
// "/twirp/twitch.example.Haberdasher/").
func pathPrefix(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	return fmt.Sprintf("/twirp/%s/", fullServiceName(file, service))
}

// pathFor returns the complete path for requests to a particular method on a
// particular service.
func pathFor(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) string {
	return pathPrefix(file, service) + stringutils.CamelCase(method.GetName())
}

func (t *Server) generateServerMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	}

func (t *twirp) generateServerJSONMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	servStruct := serviceStruct(service)
	methName := stringutils.CamelCase(method.GetName())
	t.P(`func (s *`, servStruct, `) serve`, methName, `JSON(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, req *`, t.pkgs["http"], `.Request) {`)
	t.P(`  var err error`)
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithMethodName(ctx, "`, methName, `")`)
	t.P(`  ctx, err = callRequestRouted(ctx, s.hooks)`)
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  defer closebody(req.Body)`)
	t.P(`  reqContent := new(`, t.goTypeName(method.GetInputType()), `)`)
	t.P(`  unmarshaler := `, t.pkgs["jsonpb"], `.Unmarshaler{AllowUnknownFields: true}`)
	t.P(`  if err = unmarshaler.Unmarshal(req.Body, reqContent); err != nil {`)
	t.P(`    err = wrapErr(err, "failed to parse request json")`)
	t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalErrorWith(err))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  // Call service method`)
	t.P(`  var respContent *`, t.goTypeName(method.GetOutputType()))
	t.P(`  func() {`)
	t.P(`    defer func() {`)
	t.P(`      // In case of a panic, serve a 500 error and then panic.`)
	t.P(`      if r := recover(); r != nil {`)
	t.P(`        s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("Internal service panic"))`)
	t.P(`        panic(r)`)
	t.P(`      }`)
	t.P(`    }()`)
	t.P(`    respContent, err = s.`, methName, `(ctx, reqContent)`)
	t.P(`  }()`)
	t.P()
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`  if respContent == nil {`)
	t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("received a nil *`, t.goTypeName(method.GetOutputType()), ` and nil error while calling `, methName, `. nil responses are not supported"))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  ctx = callResponsePrepared(ctx, s.hooks)`)
	t.P()
	t.P(`  var buf `, t.pkgs["bytes"], `.Buffer`)
	t.P(`  marshaler := &`, t.pkgs["jsonpb"], `.Marshaler{OrigName: true}`)
	t.P(`  if err = marshaler.Marshal(&buf, respContent); err != nil {`)
	t.P(`    err = wrapErr(err, "failed to marshal json response")`)
	t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalErrorWith(err))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithStatusCode(ctx, `, t.pkgs["http"], `.StatusOK)`)
	t.P(`  resp.Header().Set("Content-Type", "application/json")`)
	t.P(`  resp.WriteHeader(`, t.pkgs["http"], `.StatusOK)`)
	t.P(`  if _, err = resp.Write(buf.Bytes()); err != nil {`)
	t.P(`    `, t.pkgs["log"], `.Printf("errored while writing response to client, but already sent response status code to 200: %s", err)`)
	t.P(`  }`)
	t.P(`  callResponseSent(ctx, s.hooks)`)
	t.P(`}`)
	t.P()
}

func (t *twirp) generateServerProtobufMethod(service *descriptor.ServiceDescriptorProto, method *descriptor.MethodDescriptorProto) {
	servStruct := serviceStruct(service)
	methName := stringutils.CamelCase(method.GetName())

	t.P(`func (s *`, servStruct, `) serve`, methName, `Protobuf(ctx `, t.pkgs["context"], `.Context, resp `, t.pkgs["http"], `.ResponseWriter, req *`, t.pkgs["http"], `.Request) {`)
	t.P(`  var err error`)
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithMethodName(ctx, "`, methName, `")`)
	t.P(`  ctx, err = callRequestRouted(ctx, s.hooks)`)
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  defer closebody(req.Body)`)
	t.P(`  buf, err := `, t.pkgs["ioutil"], `.ReadAll(req.Body)`)
	t.P(`  if err != nil {`)
	t.P(`    err = wrapErr(err, "failed to read request body")`)
	t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalErrorWith(err))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`  reqContent := new(`, t.goTypeName(method.GetInputType()), `)`)
	t.P(`  if err = `, t.pkgs["proto"], `.Unmarshal(buf, reqContent); err != nil {`)
	t.P(`    err = wrapErr(err, "failed to parse request proto")`)
	t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalErrorWith(err))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  // Call service method`)
	t.P(`  var respContent *`, t.goTypeName(method.GetOutputType()))
	t.P(`  func() {`)
	t.P(`    defer func() {`)
	t.P(`      // In case of a panic, serve a 500 error and then panic.`)
	t.P(`      if r := recover(); r != nil {`)
	t.P(`        s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("Internal service panic"))`)
	t.P(`        panic(r)`)
	t.P(`      }`)
	t.P(`    }()`)
	t.P(`    respContent, err = s.`, methName, `(ctx, reqContent)`)
	t.P(`  }()`)
	t.P()
	t.P(`  if err != nil {`)
	t.P(`    s.writeError(ctx, resp, err)`)
	t.P(`    return`)
	t.P(`  }`)
	t.P(`  if respContent == nil {`)
	t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalError("received a nil *`, t.goTypeName(method.GetOutputType()), ` and nil error while calling `, methName, `. nil responses are not supported"))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  ctx = callResponsePrepared(ctx, s.hooks)`)
	t.P()
	t.P(`  respBytes, err := `, t.pkgs["proto"], `.Marshal(respContent)`)
	t.P(`  if err != nil {`)
	t.P(`    err = wrapErr(err, "failed to marshal proto response")`)
	t.P(`    s.writeError(ctx, resp, `, t.pkgs["twirp"], `.InternalErrorWith(err))`)
	t.P(`    return`)
	t.P(`  }`)
	t.P()
	t.P(`  ctx = `, t.pkgs["ctxsetters"], `.WithStatusCode(ctx, `, t.pkgs["http"], `.StatusOK)`)
	t.P(`  resp.Header().Set("Content-Type", "application/protobuf")`)
	t.P(`  resp.WriteHeader(`, t.pkgs["http"], `.StatusOK)`)
	t.P(`  if _, err = resp.Write(respBytes); err != nil {`)
	t.P(`    `, t.pkgs["log"], `.Printf("errored while writing response to client, but already sent response status code to 200: %s", err)`)
	t.P(`  }`)
	t.P(`  callResponseSent(ctx, s.hooks)`)
	t.P(`}`)
	t.P()
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
	return fmt.Sprintf("twirpFileDescriptor%d", t.filesHandled)
}

func (t *Server) generateServiceMetadataAccessors(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) {
	servStruct := serviceStruct(service)
	index := 0
	for i, s := range file.Service {
		if s.GetName() == service.GetName() {
			index = i
		}
	}
	t.P(`func (s *`, servStruct, `) ServiceDescriptor() ([]byte, int) {`)
	t.P(`  return `, t.serviceMetadataVarName(), `, `, strconv.Itoa(index))
	t.P(`}`)
	t.P()
	t.P(`func (s *`, servStruct, `) ProtocGenTwirpVersion() (string) {`)
	t.P(`  return `, strconv.Quote(gen.Version))
	t.P(`}`)
}

func (t *Server) generateFileDescriptor(file *descriptor.FileDescriptorProto) {
	// Copied straight of of protoc-gen-go, which trims out comments.
	pb := proto.Clone(file).(*descriptor.FileDescriptorProto)
	pb.SourceCodeInfo = nil

	b, err := proto.Marshal(pb)
	if err != nil {
		gen.Fail(err.Error())
	}

	var buf bytes.Buffer
	w, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	w.Write(b)
	w.Close()
	b = buf.Bytes()

	v := t.serviceMetadataVarName()
	t.P()
	t.P("var ", v, " = []byte{")
	t.P("	// ", fmt.Sprintf("%d", len(b)), " bytes of a gzipped FileDescriptorProto")
	for len(b) > 0 {
		n := 16
		if n > len(b) {
			n = len(b)
		}

		s := ""
		for _, c := range b[:n] {
			s += fmt.Sprintf("0x%02x,", c)
		}
		t.P(`	`, s)

		b = b[n:]
	}
	t.P("}")
}


// Given a protobuf name for a Message, return the Go name we will use for that
// type, including its package prefix.
func (t *Server) goTypeName(protoName string) string {
	def := t.reg.MessageDefinition(protoName)
	if def == nil {
		gen.Fail("could not find message for", protoName)
	}

	var prefix string
	if pkg := t.goPackageName(def.File); pkg != t.genPkgName {
		prefix = pkg + "."
	}

	var name string
	for _, parent := range def.Lineage() {
		name += parent.Descriptor.GetName() + "_"
	}
	name += def.Descriptor.GetName()
	return prefix + name
}

func (t *Server) goPackageName(file *descriptor.FileDescriptorProto) string {
	return t.fileToGoPackageName[file]
}

func unexported(s string) string { return strings.ToLower(s[:1]) + s[1:] }

func fullServiceName(file *descriptor.FileDescriptorProto, service *descriptor.ServiceDescriptorProto) string {
	name := stringutils.CamelCase(service.GetName())
	if pkg := pkgName(file); pkg != "" {
		name = pkg + "." + name
	}
	return name
}

func pkgName(file *descriptor.FileDescriptorProto) string {
	return file.GetPackage()
}

func serviceName(service *descriptor.ServiceDescriptorProto) string {
	return stringutils.CamelCase(service.GetName())
}

func serviceStruct(service *descriptor.ServiceDescriptorProto) string {
	return unexported(serviceName(service)) + "Server"
}

func methodName(method *descriptor.MethodDescriptorProto) string {
	return stringutils.CamelCase(method.GetName())
}

func fileDescSliceContains(slice []*descriptor.FileDescriptorProto, f *descriptor.FileDescriptorProto) bool {
	for _, sf := range slice {
		if f == sf {
			return true
		}
	}
	return false
}