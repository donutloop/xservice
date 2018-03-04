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

type FuncGeneratorMetaData struct {
	Name        string
	Lines       []string
	Returns     string
	Params      string
	MethodOfTyp string
	TypShortcut string
	Fnc         string
	Comment     string
}

type FuncGenerator struct {
	GoGenerator
	GoBlockGenerator
	TypeFuncMetadata FuncGeneratorMetaData
}

const funcTplName = "func"
const funcTpl string = `
{{ if .Comment }} // {{ .Comment }} {{ end }}
func {{ .Name }} ({{ .Params }}) {{if .Returns }} ({{- .Returns }}) {{end}} {
{{range $i, $line := .Lines}}
	{{- $line | safe }}
{{- end -}}
}`

func NewGoFunc(name string, parameters []*Parameter, returns []TypeReference, comment string) (*FuncGenerator, error) {

	gen := &FuncGenerator{}

	if name == "" {
		return nil, NewGeneratorErrorString(gen, "name of func is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	if err := ValidateParameters(parameters); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	gen.TypeFuncMetadata = FuncGeneratorMetaData{
		Name:    Identifier(name),
		Params:  paramList(parameters),
		Comment: comment,
	}

	if len(returns) > 0 {
		gen.TypeFuncMetadata.Returns = typeList(returns)
	}

	gen.TplName = funcTplName
	gen.InitTemplate(funcTpl)
	return gen, nil
}

func (gen *FuncGenerator) Render() (string, error) {
	gen.TypeFuncMetadata.Lines = gen.GoBlockGenerator.MetaData.Lines
	s, err := gen.renderAndFormat(gen.TypeFuncMetadata)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
