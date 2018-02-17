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
			s := UnsafeIdentifierList(test.input)
			if s != test.output {
				t.Errorf(`unepxected identifier list (actual: %s, expected: %s)`, s, test.output)
			}
		})
	}
}
