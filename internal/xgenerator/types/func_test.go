package types

import (
	"github.com/donutloop/xserver/generator/types"
	"testing"
)

func TestNewFuncAndRender(t *testing.T) {
	funcGenerator, err := types.NewGoFunc("Split",
		[]*types.Parameter{
			types.NewParameterWithTypeReference("s", types.String),
			types.NewParameterWithTypeReference("sep", types.String),
		},
		[]types.TypeReference{
			types.String,
			types.String,
		},
	)
	if err != nil {
		t.Error(err)
		return
	}

	renderedFunc, err := funcGenerator.Render()
	if err != nil {
		t.Error(err)
		return
	}

	expectedFunc := `func Split(s string, sep string) (string, string) {
}`
	if expectedFunc != renderedFunc {
		t.Errorf(`Unexpected func definition (Actual: "%s", Expected: "%s")`, renderedFunc, expectedFunc)
		return
	}
}
