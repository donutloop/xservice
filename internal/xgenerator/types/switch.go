package types

const typeSwitchTpl string = `
switch {{if .Typed }}{{.Typed }}:={{end}} {{.Var}}.(type) {
	{{range $i, $CaseData := .Cases -}}
		case {{ $CaseData.Type }}:
			{{range $i, $line := $CaseData.Code -}}
				{{ $line }}
			{{- end}}
	{{- end -}}
}
`

type SwitchGeneratorMetaData struct {
	Cases []Case
	Var   string
	Typed string
}

type SwitchGenerator struct {
	GoGenerator
	SwitchMetaData SwitchGeneratorMetaData
}

type Case struct {
	Type string
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

func NewCaseGenerator(typ string) (*CaseGenerator, error) {
	gen := &CaseGenerator{}
	if typ == "" {
		return nil, NewGeneratorErrorString(gen, "typ is missing")
	}
	if err := ValidateIdent(typ); err != nil {
		return nil, NewGeneratorError(gen, err)
	}
	gen.CaseMetaData.Typ = typ
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
