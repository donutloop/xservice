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

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"html"
	"reflect"
	"strings"
	"text/template"
)

const (
	CallerTpl                 string = "%s(%s)\n"
	DefAssginCallTpl          string = "%s := %s(%s)\n"
	DefCallTpl                string = "%s = %s(%s)\n"
	DefAppendTpl              string = "%s = append(%s)\n"
	DefSCallTpl               string = "%s := %s(%s)"
	DefVarShortTpl            string = "%s := %s\n"
	DefVarLongTpl             string = "var %s %s\n"
	DefNewTpl                 string = "%s := new(%s)\n"
	DefOperationTpl           string = "%s := %s %s %s\n"
	DefAssertTpl              string = "%s := %s.(%s)\n"
	DefStructTpl              string = "%s := %s.%s\n"
	DefReturnTpl              string = "return \n"
	DefReturnWithValuesTpl    string = "return %s\n"
	CommandTpl                string = "%s %s\n"
	StructAssignTpl           string = "%s.%s = %s\n"
	IfStatmentTpl             string = "if %s %s %s {\n"
	IfStatmentWithOwnScopeTpl string = "if %s; %s %s %s {\n"
	ElseStatmentTpl           string = "} else {\n"
	IfEndTpl                  string = "}\n"
	RangeTpl                  string = "for %s,%s := range %s {\n"
	RangeEndTpl               string = "}\n"
	DeferTpl                  string = "defer %s \n"
)

type GoGenerator struct {
	tpl     *template.Template
	TplName string
}

func (gen *GoGenerator) InitTemplate(tpl string) {
	funcMap := template.FuncMap{
		"safe": html.UnescapeString,
	}
	gen.tpl = template.Must(template.New(gen.TplName).Funcs(funcMap).Parse(tpl))
}

func (gen *GoGenerator) OverwriteTemplate(tpl string) {
	gen.InitTemplate(tpl)
}

func (gen *GoGenerator) renderAndFormat(metaData interface{}) (string, error) {
	rendered, err := gen.renderBytes(metaData)
	if err != nil {
		return "", err
	}

	formatedAndRendered, err := format.Source(rendered)
	if err != nil {
		return "", fmt.Errorf(
			`While the formting the source code is a error occurd (%v)
		||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||
		|||||||||||||||||||||||||||||||||||||||Source code||||||||||||||||||||||||||||||||||||||||||||
		||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||
		%s
		||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||
		||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||
		||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||`,
			err,
			rendered,
		)
	}

	return strings.TrimSpace(string(formatedAndRendered)), nil
}

func (gen *GoGenerator) RenderBytesData(metaData interface{}) ([]byte, error) {
	return gen.renderBytes(metaData)
}

func (gen *GoGenerator) render(metaData interface{}) (string, error) {
	buff, err := gen.populateBuffer(metaData)
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}

func (gen *GoGenerator) renderBytes(metaData interface{}) ([]byte, error) {
	buff, err := gen.populateBuffer(metaData)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (gen *GoGenerator) populateBuffer(metaData interface{}) (*bytes.Buffer, error) {
	buff := new(bytes.Buffer)
	if err := gen.tpl.Execute(buff, metaData); err != nil {
		return nil, err
	}

	return buff, nil
}

func (gen *GoGenerator) PrepareComment(comment string) []string {
	lines := make([]string, 0)
	words := strings.Fields(comment)
	tmp := words
	if len(tmp) > 15 {
		for {
			if len(tmp) == 0 {
				break
			}
			if len(tmp) > 15 {
				lines = append(lines, strings.Join(tmp[:15], " "))
				tmp = tmp[15:]
			} else {
				lines = append(lines, strings.Join(tmp, " "))
				tmp = make([]string, 0)
			}
		}
	} else {
		lines = append(lines, strings.Join(words, " "))
	}
	return lines
}

func (gen *GoGenerator) renderAll(elements interface{}) ([]string, error) {
	generators := gen.convertToGenerators(elements)
	rendered := make([]string, len(generators))
	for i, generator := range generators {
		if generator == nil {
			return nil, NewGeneratorErrorString(gen, fmt.Sprintf("generator is nil (%v)", generator))
		}

		s, err := generator.Render()
		if err != nil {
			return nil, NewGeneratorError(gen, err)
		}

		rendered[i] = s
	}

	return rendered, nil
}

func (gen *GoGenerator) convertToGenerators(elements interface{}) []Generator {
	typOfElements := reflect.TypeOf(elements)
	valueOfElements := reflect.ValueOf(elements)

	if !valueOfElements.IsValid() {
		panic(errors.New("elements is invalid"))
	}

	if typOfElements.Kind() != reflect.Slice {
		panic(errors.New("elements is not a slice"))
	}

	generators := make([]Generator, valueOfElements.Len())
	for i := 0; i < valueOfElements.Len(); i++ {
		generator, ok := valueOfElements.Index(i).Interface().(Generator)
		if !ok {
			panic(errors.New("element is not a generator"))
		}
		generators[i] = generator
	}

	return generators
}
