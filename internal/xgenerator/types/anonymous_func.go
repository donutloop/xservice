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

type AnonymousFuncGeneratorMetaData struct {
	Name        string
	Lines       []string
	Returns     string
	Params      string
	MethodOfTyp string
	TypShortcut string
	Fnc         string
	Comment     []string
}

type AnonymousFuncGenerator struct {
	GoGenerator
	GoBlockGenerator
	TypeFuncMetadata AnonymousFuncGeneratorMetaData
}

const AnonymousfuncTplName = "func"
const AnonymousfuncTpl string = `
{{ .Name }} := func ({{ .Params }}) {{if .Returns }} ({{- .Returns }}) {{end}} {
{{range $i, $line := .Lines}}
	{{- $line | safe }}
{{- end -}}
}`

func NewAnonymousGoFunc(varName string, parameters []*Parameter, returns []TypeReference) (*AnonymousFuncGenerator, error) {

	gen := &AnonymousFuncGenerator{}

	if varName == "" {
		return nil, NewGeneratorErrorString(gen, "name of func is missing")
	}

	if err := ValidateIdent(varName); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	if err := ValidateParameters(parameters); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	gen.TypeFuncMetadata = AnonymousFuncGeneratorMetaData{
		Name:   Identifier(varName),
		Params: paramList(parameters),
	}

	if len(returns) > 0 {
		gen.TypeFuncMetadata.Returns = typeList(returns)
	}

	gen.TplName = AnonymousfuncTplName
	gen.InitTemplate(AnonymousfuncTpl)
	return gen, nil
}

func (gen *AnonymousFuncGenerator) Render() (string, error) {
	gen.TypeFuncMetadata.Lines = gen.GoBlockGenerator.MetaData.Lines
	s, err := gen.renderAndFormat(gen.TypeFuncMetadata)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
