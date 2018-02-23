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

package types_test

import (
	"github.com/donutloop/xservice/internal/xgenerator/types"
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
