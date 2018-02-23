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

type MethodGeneratorMetaData struct {
	Name        string
	Lines       []string
	Returns     string
	Params      string
	MethodOfTyp string
	TypShortcut string
	Fnc         string
	Comment     []string
}

type MethodGenerator struct {
	GoGenerator
	GoBlockGenerator
	TypeFuncMetadata MethodGeneratorMetaData
}

const methodTplName string = "method"
const methodTpl string = `
{{if .Comment }}
{{range $i, $line := .Comment}}
// {{index $line -}}
{{end}}
{{- end}}
func ({{ .TypShortcut }} {{ .MethodOfTyp }}) {{ .Name }}({{ .Params }}) {{if .Returns }} ({{- .Returns }}) {{end}} {
{{range $i, $line := .Lines}}
{{- $line | safe }}
{{- end -}}
}`

func NewGoMethod(TypShortcut, methodOfTyp, name string, parameters []*Parameter, returns []TypeReference, comment string) (*MethodGenerator, error) {

	gen := &MethodGenerator{}
	if methodOfTyp == "" {
		return nil, NewGeneratorErrorString(gen, "method binding is missing")
	}

	if TypShortcut == "" {
		return nil, NewGeneratorErrorString(gen, "typ shortcut of method is missing")
	}

	if name == "" {
		return nil, NewGeneratorErrorString(gen, "name of method is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	if err := ValidateIdent(TypShortcut); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	if err := ValidateParameters(parameters); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	gen.TypeFuncMetadata = MethodGeneratorMetaData{
		Name:        Identifier(name),
		Params:      paramList(parameters),
		MethodOfTyp: methodOfTyp,
		TypShortcut: TypShortcut,
	}

	if len(returns) > 0 {
		gen.TypeFuncMetadata.Returns = typeList(returns)
	}

	if comment != "" {
		gen.TypeFuncMetadata.Comment = gen.PrepareComment(comment)
	}

	gen.TplName = methodTplName
	gen.InitTemplate(methodTpl)

	return gen, nil
}

func (gen *MethodGenerator) Render() (string, error) {
	gen.TypeFuncMetadata.Lines = gen.GoBlockGenerator.MetaData.Lines
	s, err := gen.renderAndFormat(gen.TypeFuncMetadata)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
