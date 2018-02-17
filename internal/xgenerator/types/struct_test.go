package types_test

import (
	"github.com/donutloop/xservice/internal/xgenerator/types"
	"testing"
)

func TestNewGoStructAndRender(t *testing.T) {

	structGenerator, err := types.NewGoStruct("Strings", true)
	if err != nil {
		t.Error(err)
		return
	}

	structGenerator.AddExportedField("raw", types.String, "")

	renderedStruct, err := structGenerator.Render()
	if err != nil {
		t.Error(err)
		return
	}

	expectedStruct := `type Strings struct {
	Raw string
}`
	if expectedStruct != renderedStruct {
		t.Errorf(`Unexpected struct definition (Actual: "%s", Expected: "%s")`, renderedStruct, expectedStruct)
		return
	}
}
