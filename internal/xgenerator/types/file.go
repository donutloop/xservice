package types

import (
	"fmt"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const packageTplName string = "package"
const packageTpl string = `// Code generated by xservice. DO NOT EDIT.
// source: {{ .SourceFile }}

{{if .HeaderComment }} // {{- .HeaderComment }} {{end}}

package {{ .Pkg }}

import (
{{range $i, $import := .Imports}}
             "{{- $import}}"
{{end}}
)

{{range $i, $const := .Consts}}
             {{- $const}}
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

type fileMetadata struct {
	Imports           []string
	FileName          string
	DirToScan         string
	SourceFile        string
	Pkg               string
	Types             []string
	Funcs             []string
	Consts            []string
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

func NewGoFile(pkg string, fileName string, dirToScan string) (*FileGenerator, error) {

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
	if dirToScan == "" {
		return nil, NewGeneratorErrorString(gen, "dir to scan of go file is missing")
	}

	gen.FileMetaData = fileMetadata{
		Pkg:       pkg,
		DirToScan: dirToScan,
	}

	var err error
	fileName, err = gen.prepareFileName(fileName)
	if err != nil {
		return nil, err
	}
	gen.FileMetaData.FileName = fileName

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

func (gen *FileGenerator) Import(Import string) error {
	if Import == "" {
		return NewGeneratorErrorString(gen, "import is a empty string")
	}

	gen.FileMetaData.Imports = append(gen.FileMetaData.Imports, Import)
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
	s, err := gen.Render()
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

func (gen *FileGenerator) prepareFileName(fileName string) (string, error) {
	filePathParts := strings.Split(fileName, string(os.PathSeparator))
	if len(filePathParts) > 1 {
		fileName, err := GoFileName(filePathParts[len(filePathParts)-1])
		if err != nil {
			return "", NewGeneratorErrorString(gen, fmt.Sprintf("file name contains invalid chars (%v)", err))
		}
		filePathParts[len(filePathParts)-1] = fileName

		return string(os.PathSeparator) + filepath.Join(filePathParts...), nil
	}

	fileName, err := GoFileName(fileName)
	if err != nil {
		return "", NewGeneratorErrorString(gen, fmt.Sprintf("file name contains invalid chars (%v)", err))
	}
	return fileName, nil
}

func (gen *FileGenerator) CreatePopulatedFile() error {

	pkgContent, err := gen.RenderBytes()
	if err != nil {
		return NewGeneratorError(gen, err)
	}

	formatedContent, err := imports.Process(gen.GetFileName(), pkgContent, nil)
	if err != nil {
		return NewGeneratorErrorString(gen, fmt.Sprintf(
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

	err = ioutil.WriteFile(gen.GetFileName(), formatedContent, os.ModePerm)
	if err != nil {
		return NewGeneratorErrorString(gen, fmt.Sprintf("error writing go file (%s)", err))
	}

	return nil
}

func (gen *FileGenerator) GetFileName() string {
	return gen.FileMetaData.FileName
}
