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

import "fmt"

type ClosureGeneratorMetaData struct {
	Name        string
	Lines       []string
	Returns     string
	Params      string
	MethodOfTyp string
	TypShortcut string
	Fnc         string
	Comment     []string
}

type ClosureGenerator struct {
	GoGenerator
	GoBlockGenerator
	TypeFuncMetadata ClosureGeneratorMetaData
}

const closureTplName = "closure"
const closureTpl string = `
func {{ .Name }} ({{ .Params }}) {{if .Returns }} ({{- .Returns }}) {{end}} {
	{{if .Lines }}
		{{range $i, $line := .Lines}}
			{{- $line}}
		{{- end}}
	{{- end}}
	{{- .Fnc }}
}`

const innerFuncTplName string = "innerFunc"
const innerFuncTpl string = `
return func ({{ .Params }}) {{if .Returns }} ({{- .Returns }}) {{end}} {
{{range $i, $line := .Lines}}
	{{- $line | safe }}
{{- end -}}
}`

func NewClosureFunc(name string, parameters []*Parameter, f *FuncGenerator) (*ClosureGenerator, error) {

	gen := &ClosureGenerator{}
	if name == "" {
		return nil, NewGeneratorErrorString(gen, "name of closure is missing")
	}

	if f == nil {
		return nil, NewGeneratorErrorString(gen, "inner func of closure is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	if err := ValidateParameters(parameters); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	f.TplName = innerFuncTplName
	f.OverwriteTemplate(innerFuncTpl)
	innerFunc, err := f.Render()
	if err != nil {
		return nil, err
	}

	gen.TypeFuncMetadata = ClosureGeneratorMetaData{
		Name:    Identifier(name),
		Params:  paramList(parameters),
		Returns: fmt.Sprintf("%s func(%s) (%s)", f.TypeFuncMetadata.Name, f.TypeFuncMetadata.Params, f.TypeFuncMetadata.Returns),
		Fnc:     innerFunc,
	}

	gen.TplName = closureTplName
	gen.InitTemplate(closureTpl)
	return gen, nil
}

func (gen *ClosureGenerator) Render() (string, error) {
	gen.TypeFuncMetadata.Lines = gen.GoBlockGenerator.MetaData.Lines
	s, err := gen.renderAndFormat(gen.TypeFuncMetadata)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
