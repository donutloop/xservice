package types

// Template of const
const constTpl string = `
{{ if .Comment }} // {{ .Comment }} {{ end }}
const {{ .Name }} {{ .Typ }} = {{ .Value }}`
const constTplName string = "const"

type ConstMetadata struct {
	Name  string
	Value string
	Typ   string
	Comment string
}

type ConstGenerator struct {
	GoGenerator
	ConstMetaData ConstMetadata
}

func NewGoConst(name string, typeReference TypeReference, value string, commentGenerator *CommentGenerator) (*ConstGenerator, error) {

	constGen := ConstGenerator{}

	if name == "" {
		return nil, NewGeneratorErrorString(constGen, "name of const is empty")
	}

	if err := ValidateIdent(name); err != nil {
		return nil, NewGeneratorError(constGen, err)
	}

	if value == "" {
		return nil, NewGeneratorErrorString(constGen, "value of const is missing")
	}

	if typeReference == nil {
		return nil, NewGeneratorErrorString(constGen, "TypeReference is missing")

	}

	if typeReference.GetName() == "" {
		return nil, NewGeneratorErrorString(constGen, "type is missing")
	}

	constGen.ConstMetaData = ConstMetadata{
		Name:  name,
		Value: value,
		Typ:   typeReference.GetName(),
	}

	if commentGenerator != nil {
		comment, err := commentGenerator.Render()
		if err != nil {
			return nil, err
		}

		if comment != "" {
			constGen.ConstMetaData.Comment = comment
		}
	}

	constGen.TplName = constTplName
	constGen.InitTemplate(constTpl)

	return &constGen, nil
}

func (gen *ConstGenerator) Render() (string, error) {
	s, err := gen.renderAndFormat(gen.ConstMetaData)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
