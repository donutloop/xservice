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

func TestNewGoConstAndRender(t *testing.T) {

	constGenerator, err := types.NewGoConst("dummy", types.String, `"value"`, nil)
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
