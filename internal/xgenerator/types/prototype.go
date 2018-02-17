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
