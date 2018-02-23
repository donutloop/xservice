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

type InitStructFieldTmplValues struct {
	Fields []*InitStructField
	Typ    string
}

type InitStructField struct {
	Name  string
	Value string
}

type InitStructGenerator struct {
	GoGenerator
	MetaData InitStructFieldTmplValues
}

const initStructTplName string = "initStruct"
const initStructTpl string = `
{{.Typ}} {
	{{range .Fields}}
        {{- .Name}}: {{.Value}},
	{{end}}
}`

func NewInitGoStruct(typ string) (*InitStructGenerator, error) {
	initStructGenerator := InitStructGenerator{}
	initStructGenerator.TplName = initStructTplName
	initStructGenerator.InitTemplate(initStructTpl)
	initStructGenerator.MetaData.Typ = typ
	return &initStructGenerator, nil
}

func (gen *InitStructGenerator) AddUnexportedValueToField(name, value string) error {
	return gen.addValueToField(UnexportedIdentifier(name), value)
}

func (gen *InitStructGenerator) AddExportedValueToField(name, value string) error {
	return gen.addValueToField(ExportedIdentifier(name), value)
}

func (gen *InitStructGenerator) addValueToField(name, value string) error {

	if name == "" {
		return NewGeneratorErrorString(gen, "unexported field name is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return NewGeneratorError(gen, err)
	}

	if err := ValidateIdent(value); err != nil {
		return NewGeneratorError(gen, err)
	}

	field := &InitStructField{
		Name:  name,
		Value: value,
	}

	gen.MetaData.Fields = append(gen.MetaData.Fields, field)

	return nil
}

func (gen *InitStructGenerator) Render() (string, error) {
	s, err := gen.render(gen.MetaData)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
