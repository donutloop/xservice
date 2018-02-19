package types

const interfaceNameTpl string = "interface"

const interfaceTpl string = `
{{ if .HeaderComment }} // .HeaderComment {{ .HeaderComment }}

type {{ .Name }} interface {
{{range $i, $Prototype := .Prototypes}}
			{{if index $Prototype.Comment }}
				{{range $i, $line := index $Prototype.Comment }}
				// {{ index $line }}
				{{- end}}
			{{- end}}
            {{ index $Prototype.Name }}({{- index $Prototype.Params }}) {{if index $Prototype.Returns }} ({{- index $Prototype.Returns }}) {{end}}
{{end}}
}`

type InterfaceMetadata struct {
	Name       string
	HeaderComment string
	Prototypes []InterfacePrototypeMetadata
}

type InterfacePrototypeMetadata struct {
	Name    string
	Params  string
	Returns string
	Comment []string
}

type InterfaceGenerator struct {
	GoGenerator
	InterfaceMetadata InterfaceMetadata
}

func NewGoInterface(name string) (*InterfaceGenerator, error) {

	interfaceGenerator := InterfaceGenerator{}
	if name == "" {
		return nil, NewGeneratorErrorString(interfaceGenerator, "name of interface is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return nil, NewGeneratorError(interfaceGenerator, err)
	}

	interfaceGenerator.InterfaceMetadata = InterfaceMetadata{
		Name: ExportedIdentifier(name),
	}

	interfaceGenerator.TplName = interfaceNameTpl
	interfaceGenerator.InitTemplate(interfaceTpl)
	return &interfaceGenerator, nil
}

func (gen *InterfaceGenerator) Prototype(name string, parameters []*Parameter, returns []TypeReference, comment string) error {
	if name == "" {
		return NewGeneratorErrorString(gen, "name of prototype is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return NewGeneratorError(gen, err)
	}

	if err := ValidateParameters(parameters); err != nil {
		return NewGeneratorError(gen, err)
	}

	prototype := InterfacePrototypeMetadata{
		Name:    ExportedIdentifier(name),
		Params:  paramList(parameters),
		Returns: typeList(returns),
	}

	if comment != "" {
		prototype.Comment = gen.PrepareComment(comment)
	}

	gen.InterfaceMetadata.Prototypes = append(gen.InterfaceMetadata.Prototypes, prototype)

	return nil
}

func (gen *InterfaceGenerator) Render() (string, error) {
	s, err := gen.renderAndFormat(gen.InterfaceMetadata)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
