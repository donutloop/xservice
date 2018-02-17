package types

import (
	"github.com/donutloop/xserver/generator/types"
	"testing"
)

func TestNewGoConstAndRender(t *testing.T) {

	constGenerator, err := types.NewGoConst("dummy", types.String, `"value"`)
	if err != nil {
		t.Error(err)
		return
	}

	renderedConst, err := constGenerator.Render()
	if err != nil {
		t.Error(err)
		return
	}

	expectedConst := `const dummy string = "value"`
	if expectedConst != renderedConst {
		t.Errorf(`Unexpected const definition (Actual: "%s", Expected: "%s")`, renderedConst, expectedConst)
		return
	}
}
