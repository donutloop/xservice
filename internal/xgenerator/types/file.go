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

package types

import (
	"fmt"
	"golang.org/x/tools/imports"
	"os"
	"path/filepath"
	"strings"
)

const packageTplName string = "package"
const packageTpl string = `

{{if .HeaderComment }}
	{{ .HeaderComment }}
{{end}}

package {{ .Pkg }}

import (
{{range $i, $ip := .Imports}}
         {{if $ip.Alias }} {{$ip.Alias}} {{end}} "{{- $ip.ImportPath}}"
{{end}}
)

{{range $i, $const := .Consts}}
             {{- $const}}
{{end}}

{{range $i, $var := .Vars}}
            {{- $var}}
{{end}}

{{range $i, $Interface := .Interfaces}}
            {{- $Interface}}
{{end}}

{{range $i, $Prototype := .Prototypes}}
            {{- $Prototype}}
{{end}}

{{if .Types }}
type({{range $i, $typ := .Types}}
            {{- $typ}}
{{end}})
{{end}}

{{if .TypesWithoutGroup }}
	{{range $i, $typ := .TypesWithoutGroup}}
            {{- $typ}}
	{{end}}
{{end}}

{{range $i, $TypeWithMethods := .TypesWithMethods}}
            {{ $TypeWithMethods}}
{{end}}

{{range $i, $func := .Funcs}}
            {{- $func}}
{{end}}
`

type Type interface {
	Render() (string, error)
	GetMethods() []*MethodGenerator
}

type Generator interface {
	Render() (string, error)
}

type ImportDecl struct {
	Alias      string
	ImportPath string
}

type fileMetadata struct {
	Imports           []ImportDecl
	FileName          string
	DirToScan         string
	SourceFile        string
	Pkg               string
	Types             []string
	Funcs             []string
	Consts            []string
	Vars              []string
	Prototypes        []string
	Interfaces        []string
	TypesWithMethods  []string
	TypesWithoutGroup []string
	HeaderComment     string
}

type FileGenerator struct {
	GoGenerator
	FileMetaData fileMetadata
}

func NewGoFile(pkg string, fileName string) (*FileGenerator, error) {

	gen := FileGenerator{}
	if pkg == "" {
		return nil, NewGeneratorErrorString(gen, "pkg of go file is missing")
	}

	if err := ValidateIdent(pkg); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	if fileName == "" {
		return nil, NewGeneratorErrorString(gen, "filename of go file is missing")
	}

	gen.FileMetaData = fileMetadata{
		Pkg: pkg,
	}

	gen.FileMetaData.FileName = gen.prepareFileName(fileName)

	gen.TplName = packageTplName
	gen.InitTemplate(packageTpl)
	return &gen, nil
}

func (gen *FileGenerator) Type(typs ...*StructGenerator) error {

	rendered, err := gen.GoGenerator.renderAll(typs)
	if err != nil {
		return err
	}
	gen.FileMetaData.Types = append(gen.FileMetaData.Types, rendered...)

	return nil
}

func (gen *FileGenerator) HeaderComment(generator *CommentGenerator) error {

	comment, err := generator.Render()
	if err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.FileMetaData.HeaderComment = comment
	return nil
}

func (gen *FileGenerator) Import(alias string, Import string) error {
	if Import == "" {
		return NewGeneratorErrorString(gen, "import is a empty string")
	}

	i := ImportDecl{
		ImportPath: Import,
		Alias:      alias,
	}

	gen.FileMetaData.Imports = append(gen.FileMetaData.Imports, i)
	return nil
}

func (gen *FileGenerator) TypesWithoutGroup(typs ...*StructGenerator) error {

	rendered, err := gen.renderAll(typs)
	if err != nil {
		return err
	}

	gen.FileMetaData.TypesWithoutGroup = append(gen.FileMetaData.TypesWithoutGroup, rendered...)

	return nil
}

func (gen *FileGenerator) Func(fncs ...*FuncGenerator) error {

	rendered, err := gen.renderAll(fncs)
	if err != nil {
		return err
	}

	gen.FileMetaData.Funcs = append(gen.FileMetaData.Funcs, rendered...)
	return nil
}

func (gen *FileGenerator) Closure(colsures ...*ClosureGenerator) error {

	rendered, err := gen.renderAll(colsures)
	if err != nil {
		return err
	}

	gen.FileMetaData.Funcs = append(gen.FileMetaData.Funcs, rendered...)
	return nil
}

func (gen *FileGenerator) Const(g *ConstGenerator) error {

	if g == nil {
		return NewGeneratorErrorString(gen, "generator is nil")
	}

	cnst, err := g.Render()
	if err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.FileMetaData.Consts = append(gen.FileMetaData.Consts, cnst)
	return nil
}

func (gen *FileGenerator) Var(varObject string) error {
	gen.FileMetaData.Vars = append(gen.FileMetaData.Vars, varObject)
	return nil
}

func (gen *FileGenerator) Prototype(g *PrototypeGenerator) error {

	if g == nil {
		return NewGeneratorErrorString(gen, "generator is nil")
	}

	prototype, err := g.Render()
	if err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.FileMetaData.Prototypes = append(gen.FileMetaData.Prototypes, prototype)
	return nil
}

func (gen *FileGenerator) Interface(g *InterfaceGenerator) error {

	if g == nil {
		return NewGeneratorErrorString(gen, "generator is nil")
	}

	i, err := g.Render()
	if err != nil {
		return NewGeneratorError(gen, err)
	}

	gen.FileMetaData.Interfaces = append(gen.FileMetaData.Interfaces, i)
	return nil
}

func (gen *FileGenerator) TypesWithMethods(tg Type) error {

	t, err := tg.Render()
	if err != nil {
		return NewGeneratorError(gen, err)
	}

	methods := make([]string, 0, len(tg.GetMethods()))
	for _, method := range tg.GetMethods() {
		m, err := method.Render()
		if err != nil {
			return NewGeneratorError(gen, err)
		}
		methods = append(methods, m)
	}

	gen.FileMetaData.TypesWithMethods = append(gen.FileMetaData.TypesWithMethods, t)
	gen.FileMetaData.TypesWithMethods = append(gen.FileMetaData.TypesWithMethods, methods...)
	return nil
}

func (gen *FileGenerator) Render() (string, error) {
	s, err := gen.render(gen.FileMetaData)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, nil
}

func (gen *FileGenerator) RenderBytes() ([]byte, error) {
	b, err := gen.renderBytes(gen.FileMetaData)
	if err != nil {
		return b, NewGeneratorError(gen, err)
	}
	return b, err
}

func (gen *FileGenerator) prepareFileName(fileName string) string {
	filePathParts := strings.Split(fileName, string(os.PathSeparator))
	if len(filePathParts) > 1 {
		fileName = GoFileName(filePathParts[len(filePathParts)-1])
		filePathParts[len(filePathParts)-1] = fileName
		return string(os.PathSeparator) + filepath.Join(filePathParts...)
	}
	fileName = GoFileName(fileName)
	return fileName
}

func (gen *FileGenerator) RenderAndFormatCode() ([]byte, error) {

	pkgContent, err := gen.RenderBytes()
	if err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	formatedContent, err := imports.Process(gen.GetFileName(), pkgContent, nil)
	if err != nil {
		return nil, NewGeneratorErrorString(gen, fmt.Sprintf(
			`While the formting the source code is a error occurd (%v)
		||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||
		|||||||||||||||||||||||||||||||||||||||Source code||||||||||||||||||||||||||||||||||||||||||||
		||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||
		%s
		||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||
		||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||
		||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||`,
			err,
			pkgContent,
		))
	}

	return formatedContent, err
}

func (gen *FileGenerator) GetFileName() string {
	return gen.FileMetaData.FileName
}
