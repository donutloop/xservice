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

const structWithoutPkgAndGroupTplName string = "structWithoutPkgAndGroup"
const structWithoutPkgAndGroupTpl string = `
{{if .Comment }}
{{range $i, $line := .Comment}}
// {{index $line }}
{{- end}}
{{- end}}
{{.Name}} struct {
	{{range .Fields}}
        {{- .Name}} {{.Type}} {{.TagMeta | safe }}  {{if .CommentOfProperty }}// {{.CommentOfProperty }}{{end}}
	{{end -}}
}`

const structWithoutPkgTplName string = "structWithoutPkg"
const structWithoutPkgTpl string = `
{{if .Comment }}
{{range $i, $line := .Comment}}
// {{index $line }}
{{- end}}
{{- end}}
type {{.Name}} struct {
	{{range .Fields}}
        {{- .Name}} {{.Type}} {{.TagMeta | safe }} {{if .CommentOfProperty }}// {{.CommentOfProperty }}{{end}}
	{{end}}
}`

type StructTmplValues struct {
	Name    string
	Comment []string
	Fields  []*StructFieldTmplValues
	Methods []*MethodGenerator
}

type StructFieldTmplValues struct {
	Name              string
	Type              string
	TagMeta           string
	CommentOfProperty string
}

type StructGenerator struct {
	GoGenerator
	Group          bool
	StructMetaData StructTmplValues
}

func NewGoStruct(name string, withoutGroup bool, exported bool) (*StructGenerator, error) {

	structGenerator := StructGenerator{}
	if name == "" {
		return nil, NewGeneratorErrorString(structGenerator, "name of struct is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return nil, NewGeneratorError(structGenerator, err)
	}

	structGenerator.StructMetaData = StructTmplValues{}

	if exported {
		structGenerator.StructMetaData.Name = ExportedIdentifier(name)
	} else {
		structGenerator.StructMetaData.Name = UnexportedIdentifier(name)
	}

	if withoutGroup {
		structGenerator.TplName = structWithoutPkgTplName
		structGenerator.InitTemplate(structWithoutPkgTpl)
	} else {
		structGenerator.TplName = structWithoutPkgAndGroupTplName
		structGenerator.InitTemplate(structWithoutPkgAndGroupTpl)
	}

	structGenerator.Group = withoutGroup

	return &structGenerator, nil
}

func (gen *StructGenerator) Type(typ TypeReference, comment string) error {

	if typ == nil {
		return NewGeneratorErrorString(gen, "TypeReference is missing")
	}

	if typ.GetName() == "" {
		return NewGeneratorErrorString(gen, "typ of exported field is missing")
	}

	field := &StructFieldTmplValues{
		Type:              typ.GetName(),
		CommentOfProperty: comment,
	}

	gen.StructMetaData.Fields = append(gen.StructMetaData.Fields, field)
	return nil
}

func (gen *StructGenerator) AddExportedField(name string, typ TypeReference, comment string) error {

	if name == "" {
		return NewGeneratorErrorString(gen, "exported field name is missing")
	}

	if typ == nil {
		return NewGeneratorErrorString(gen, "TypeReference is missing")
	}

	if typ.GetName() == "" {
		return NewGeneratorErrorString(gen, "typ of exported field is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return NewGeneratorError(gen, err)
	}

	field := &StructFieldTmplValues{
		Name:              ExportedIdentifier(name),
		Type:              typ.GetName(),
		CommentOfProperty: comment,
	}

	gen.StructMetaData.Fields = append(gen.StructMetaData.Fields, field)
	return nil
}

func (gen *StructGenerator) Composition(typ TypeReference) error {

	if typ == nil {
		return NewGeneratorErrorString(gen, "TypeReference is missing")
	}

	if typ.GetName() == "" {
		return NewGeneratorErrorString(gen, "typ of exported field is missing")
	}

	field := &StructFieldTmplValues{
		Type: typ.GetName(),
	}

	gen.StructMetaData.Fields = append(gen.StructMetaData.Fields, field)
	return nil
}

func (gen *StructGenerator) AddUnexportedField(name string, typ TypeReference, comment string) error {

	if name == "" {
		return NewGeneratorErrorString(gen, "unexported field name is missing")
	}

	if typ == nil {
		return NewGeneratorErrorString(gen, "TypeReference is missing")
	}

	if typ.GetName() == "" {
		return NewGeneratorErrorString(gen, "typ of unexported field is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return NewGeneratorError(gen, err)
	}

	field := &StructFieldTmplValues{
		Name:              UnexportedIdentifier(name),
		Type:              typ.GetName(),
		CommentOfProperty: comment,
	}

	gen.StructMetaData.Fields = append(gen.StructMetaData.Fields, field)

	return nil
}

func (gen *StructGenerator) AddMethod(fg ...*MethodGenerator) {
	gen.StructMetaData.Methods = append(gen.StructMetaData.Methods, fg...)
}

func (gen *StructGenerator) GetMethods() []*MethodGenerator {
	return gen.StructMetaData.Methods
}

func (gen *StructGenerator) Render() (string, error) {
	if !gen.Group {
		s, err := gen.render(gen.StructMetaData)
		if err != nil {
			return s, NewGeneratorError(gen, err)
		}
		return s, err
	}
	s, err := gen.renderAndFormat(gen.StructMetaData)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
