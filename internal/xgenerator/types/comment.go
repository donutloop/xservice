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

import "fmt"

// Template of comment
const commentTpl string = `
{{if .Comment }}
	{{range $i, $line := .Comment}}
             // {{- $line -}}
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
