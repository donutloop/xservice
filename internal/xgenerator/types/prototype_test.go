package types

import (
	"github.com/donutloop/xserver/generator/types"
	"testing"
)

func TestNewGoPrototypeAndRender(t *testing.T) {

	prototypeGenerator, err := types.NewGoFuncPrototype(
		"split",
		[]*types.Parameter{
			types.NewParameterWithTypeReference("s", types.String),
			types.NewParameterWithTypeReference("sep", types.String),
		},
		[]types.TypeReference{
			types.String,
			types.String,
		},
		"",
	)

	if err != nil {
		t.Error(err)
		return
	}

	renderedPrototype, err := prototypeGenerator.Render()
	if err != nil {
		t.Error(err)
		return
	}

	expectedPrototype := "type Split func(s string, sep string) (string, string)"
	if expectedPrototype != renderedPrototype {
		t.Errorf(`Unexpected prototype definition (Actual: "%s", Expected: "%s")`, renderedPrototype, expectedPrototype)
		return
	}
}
