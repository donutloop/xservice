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
