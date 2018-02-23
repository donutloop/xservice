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

import "strconv"

// Template of slice literal
const sliceLiteralTpl string = `
{{ .Name }} := [{{ if .Len }}{{ .Len }}{{ end }}]{{ .Typ }}{
	{{range $i, $Value := .Values}}
            {{- $Value}},
	{{end}}
}`
const sliceLiteralTplName string = "sliceLiteral"

type SliceLiteralMetadata struct {
	Name   string
	Values []string
	Typ    string
	Len    string
}

type SliceLiteralGenerator struct {
	GoGenerator
	SliceMetaData SliceLiteralMetadata
}

func NewGoSliceLiteral(varName string, typeReference TypeReference, len int) (*SliceLiteralGenerator, error) {

	sliceGen := SliceLiteralGenerator{}

	if varName == "" {
		return nil, NewGeneratorErrorString(sliceGen, "name of sliceliteral is empty")
	}

	if err := ValidateIdent(varName); err != nil {
		return nil, NewGeneratorError(sliceGen, err)
	}

	if typeReference == nil {
		return nil, NewGeneratorErrorString(sliceGen, "TypeReference is missing")

	}

	if typeReference.GetName() == "" {
		return nil, NewGeneratorErrorString(sliceGen, "type is missing")
	}

	sliceGen.SliceMetaData = SliceLiteralMetadata{
		Name: varName,
		Typ:  typeReference.GetName(),
		Len:  strconv.Itoa(len),
	}

	sliceGen.TplName = sliceLiteralTplName
	sliceGen.InitTemplate(sliceLiteralTpl)

	return &sliceGen, nil
}

func (gen *SliceLiteralGenerator) Append(s string) {
	gen.SliceMetaData.Values = append(gen.SliceMetaData.Values, s)
}

func (gen *SliceLiteralGenerator) Render() (string, error) {
	s, err := gen.renderAndFormat(gen.SliceMetaData)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
