package types

type FuncGeneratorMetaData struct {
	Name        string
	Lines       []string
	Returns     string
	Params      string
	MethodOfTyp string
	TypShortcut string
	Fnc         string
	Comment     []string
}

type FuncGenerator struct {
	GoGenerator
	GoBlockGenerator
	TypeFuncMetadata FuncGeneratorMetaData
}

const funcTplName = "func"
const funcTpl string = `
func {{ .Name }} ({{ .Params }}) {{if .Returns }} ({{- .Returns }}) {{end}} {
{{range $i, $line := .Lines}}
	{{- $line | safe }}
{{- end -}}
}`

func NewGoFunc(name string, parameters []*Parameter, returns []TypeReference) (*FuncGenerator, error) {

	gen := &FuncGenerator{}

	if name == "" {
		return nil, NewGeneratorErrorString(gen, "name of func is missing")
	}

	if err := ValidateIdent(name); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	if err := ValidateParameters(parameters); err != nil {
		return nil, NewGeneratorError(gen, err)
	}

	gen.TypeFuncMetadata = FuncGeneratorMetaData{
		Name:   Identifier(name),
		Params: paramList(parameters),
	}

	if len(returns) > 0 {
		gen.TypeFuncMetadata.Returns = typeList(returns)
	}

	gen.TplName = funcTplName
	gen.InitTemplate(funcTpl)
	return gen, nil
}

func (gen *FuncGenerator) Render() (string, error) {
	gen.TypeFuncMetadata.Lines = gen.GoBlockGenerator.MetaData.Lines
	s, err := gen.renderAndFormat(gen.TypeFuncMetadata)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
