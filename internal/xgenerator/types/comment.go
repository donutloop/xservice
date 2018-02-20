package types

import "fmt"

// Template of const
const commentTpl string = `
{{if .Comment }}
	{{range $i, $line := .Comment}}
             // {{- $line}}
	{{end}}
{{end}}
`
const commentTplName string = "const"

type CommentMetaData struct {
	Comment []string
}

type CommentGenerator struct {
	GoGenerator
	CommentMetaData CommentMetaData
}

func NewGoComment() *CommentGenerator {

	commentGen := new(CommentGenerator)
	commentGen.TplName = commentTplName
	commentGen.InitTemplate(commentTpl)

	return commentGen
}

func (gen *CommentGenerator) P(s string) {
	gen.CommentMetaData.Comment = append(gen.CommentMetaData.Comment, s)
}

func (gen *CommentGenerator) Pf(format string, a ...interface{}) {
	gen.CommentMetaData.Comment = append(gen.CommentMetaData.Comment, fmt.Sprintf(format, a))
}

func (gen *CommentGenerator) Render() (string, error) {
	s, err := gen.renderAndFormat(gen.CommentMetaData)
	if err != nil {
		return s, NewGeneratorError(gen, err)
	}
	return s, err
}
