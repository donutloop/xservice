package types

import (
	"github.com/donutloop/xserver/generator/types"
	"testing"
)

func TestNewInterfaceAndRender(t *testing.T) {
	interfaceGenerator, err := types.NewGoInterface("Stringer")
	if err != nil {
		t.Error(err)
		return
	}

	err = interfaceGenerator.Prototype("split",
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

	renderedInterface, err := interfaceGenerator.Render()
	if err != nil {
		t.Error(err)
		return
	}

	expectedInterface := `type Stringer interface {
	Split(s string, sep string) (string, string)
}`
	if expectedInterface != renderedInterface {
		t.Errorf(`Unexpected interface definition (Actual: "%s", Expected: "%s")`, renderedInterface, expectedInterface)
		return
	}
}
