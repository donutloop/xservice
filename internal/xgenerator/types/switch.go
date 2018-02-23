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

const typeSwitchTpl string = `
switch {{if .Typed }}{{.Typed }}:={{end}} {{.Var}}{{if .Typed }}.(type){{end}} {
	{{range $i, $CaseData := .Cases -}}
		case {{ $CaseData.Type }}:
			{{range $i, $line := $CaseData.Code -}}
				{{ $line }}
			{{- end}}
	{{- end -}}
	{{ if .DefaultCase }}
		default:
		{{range $i, $line := .DefaultCase.Code -}}
				{{ $line }}
		{{- end -}}
	{{- end -}}
}
`

type SwitchGeneratorMetaData struct {
	Cases       []Case
	DefaultCase DefaultCase
	Var         string
	Typed       string
}

type SwitchGenerator struct {
	GoGenerator
	SwitchMetaData SwitchGeneratorMetaData
}

type Case struct {
	Type string
	Code []string
}

type DefaultCase struct {
	Code []string
}

func NewSwitchGenerator(varName string) (*SwitchGenerator, error) {
	gen := &SwitchGenerator{}
	if varName == "" {
		return nil, NewGeneratorErrorString(gen, "var name is missing")
	}
	if err := ValidateIdent(varName); err != nil {
		return nil, NewGeneratorError(gen, err)
	}
	gen.SwitchMetaData.Var = varName
	gen.TplName = "type_switch"
	gen.InitTemplate(typeSwitchTpl)
	return gen, nil
}

func (gen *SwitchGenerator) UseAssertion(ok bool) {
	if ok {
		gen.SwitchMetaData.Typed = "o"
		return
	}
}

func (gen *SwitchGenerator) Case(caseGenerator CaseGenerator) {
	gen.SwitchMetaData.Cases = append(gen.SwitchMetaData.Cases, Case{Type: caseGenerator.CaseMetaData.Typ, Code: caseGenerator.GoBlockGenerator.MetaData.Lines})
}

func (gen *SwitchGenerator) Default(caseDefaultGenerator DefaultCaseGenerator) {
	gen.SwitchMetaData.DefaultCase = DefaultCase{Code: caseDefaultGenerator.GoBlockGenerator.MetaData.Lines}
}

func (gen *SwitchGenerator) Render() (string, error) {
	s, err := gen.render(gen.SwitchMetaData)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}

type CaseGeneratorMetaData struct {
	Lines []string
	Typ   string
}

type CaseGenerator struct {
	GoGenerator
	GoBlockGenerator
	CaseMetaData CaseGeneratorMetaData
}

const CaseTpl string = `
case {{ .Typ }}:
{{range $i, $line := .Lines -}}
	{{ $line }}
{{- end}}
`

func NewCaseGenerator(value string) (*CaseGenerator, error) {
	gen := &CaseGenerator{}
	if value == "" {
		return nil, NewGeneratorErrorString(gen, "value is missing")
	}
	if err := ValidateIdent(value); err != nil {
		return nil, NewGeneratorError(gen, err)
	}
	gen.CaseMetaData.Typ = value
	gen.TplName = "case"
	gen.InitTemplate(CaseTpl)
	return gen, nil
}

func (gen *CaseGenerator) Render() (string, error) {
	gen.CaseMetaData.Lines = gen.GoBlockGenerator.MetaData.Lines
	s, err := gen.render(gen.CaseMetaData)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}

const DefaultCaseTpl string = `
default:
{{range $i, $line := .Lines -}}
	{{ $line }}
{{- end}}
`

type DefaultCaseGeneratorMetaData struct {
	Lines []string
}

type DefaultCaseGenerator struct {
	GoGenerator
	GoBlockGenerator
	CaseMetaData DefaultCaseGeneratorMetaData
}

func NewDefaultCaseGenerator() (*DefaultCaseGenerator, error) {
	gen := &DefaultCaseGenerator{}
	gen.TplName = "defaultCase"
	gen.InitTemplate(DefaultCaseTpl)
	return gen, nil
}

func (gen *DefaultCaseGenerator) Render() (string, error) {
	gen.CaseMetaData.Lines = gen.GoBlockGenerator.MetaData.Lines
	s, err := gen.render(gen.CaseMetaData)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
