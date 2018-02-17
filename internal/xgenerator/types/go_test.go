package types

import (
	"strings"
	"testing"
)

func newStubGenerator() *stubGenerator {
	return new(stubGenerator)
}

type stubGenerator struct{}

func (gen *stubGenerator) Render() (string, error) {
	return "rendered", nil
}

func TestGoGenerator_RenderAll(t *testing.T) {
	gen := &GoGenerator{}

	tests := []struct {
		name   string
		input  interface{}
		output string
	}{
		{
			name:   "3 Generators",
			input:  []*stubGenerator{newStubGenerator(), newStubGenerator(), newStubGenerator()},
			output: "rendered,rendered,rendered",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rendered, err := gen.renderAll(test.input)
			if err != nil {
				t.Fatal(err)
			}
			s := strings.Join(rendered, ",")

			if s != test.output {
				t.Errorf(`unexpected value (actual: "%s", expected: "%s")`, s, test.output)
			}
		})
	}
}
func TestGoGenerator_RenderAllPanic(t *testing.T) {
	gen := &GoGenerator{}

	tests := []struct {
		name       string
		input      interface{}
		errMessage string
	}{
		{
			name:       "string slice",
			input:      []string{""},
			errMessage: "element is not a generator",
		},
		{
			name:       "string",
			input:      "",
			errMessage: "elements is not a slice",
		},
		{
			name:       "nil",
			input:      nil,
			errMessage: "elements is invalid",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				v := recover()
				err, ok := v.(error)
				if !ok {
					t.Fatalf("value isn't a error (%v)", err)
				}

				if err.Error() != test.errMessage {
					t.Fatal(err)
				}
			}()

			_, err := gen.renderAll(test.input)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
