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

const prototypeTplName string = "prototype"
const prototypeTpl string = `
{{if .Comment }}
{{range $i, $line := .Comment}}
// {{index $line }}
{{- end}}
{{- end}}
type {{ .Name }} func({{ .Params }}) ({{ .Returns }})`

type PrototypeMetadata struct {
	Name    string
	Params  string
	Returns string
	Comment []string
}

type PrototypeGenerator struct {
	GoGenerator
	PrototypeMetadata PrototypeMetadata
}

func NewGoFuncPrototype(name string, parameters []*Parameter, returns []TypeReference, comment string) (*PrototypeGenerator, error) {

	prototypeGenerator := PrototypeGenerator{}
	if name == "" {
		return nil, NewGeneratorErrorString(prototypeGenerator, "name of prototype is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return nil, NewGeneratorError(prototypeGenerator, err)
	}

	if err := ValidateParameters(parameters); err != nil {
		return nil, NewGeneratorError(prototypeGenerator, err)
	}

	gen := &PrototypeGenerator{
		PrototypeMetadata: PrototypeMetadata{
			Name:    ExportedIdentifier(name),
			Params:  paramList(parameters),
			Returns: typeList(returns),
		},
	}
	if comment != "" {
		gen.PrototypeMetadata.Comment = gen.PrepareComment(comment)
	}
	gen.TplName = prototypeTplName
	gen.InitTemplate(prototypeTpl)

	return gen, nil
}

func (gen *PrototypeGenerator) Render() (string, error) {
	s, err := gen.renderAndFormat(gen.PrototypeMetadata)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
