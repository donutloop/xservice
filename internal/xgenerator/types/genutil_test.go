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

import "testing"

func TestUnsafeIdentifierList(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output string
	}{
		{
			name:   "3 identifier",
			input:  []string{"id1", "id2", "id3"},
			output: "id1, id2, id3",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := ValueList(test.input)
			if s != test.output {
				t.Errorf(`unepxected identifier list (actual: %s, expected: %s)`, s, test.output)
			}
		})
	}
}
