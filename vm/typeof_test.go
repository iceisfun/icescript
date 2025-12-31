package vm

import (
	"testing"

	"github.com/iceisfun/icescript/object"
)

func TestTypeof(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"typeof(1)", "integer"},
		{"typeof(1.5)", "float"},
		{"typeof(true)", "boolean"},
		{"typeof(false)", "boolean"},
		{`typeof("s")`, "string"},
		{"typeof(null)", "null"},
		{"typeof([1])", "array"},
		{"typeof(len)", "builtin"},
		{"typeof(testMultiReturn(1, 2))", "tuple"},
	}

	for _, tt := range tests {
		val := testEval(t, tt.input)
		strVal, ok := val.(*object.String)
		if !ok {
			t.Errorf("test input %q: expected string, got %T (%+v)", tt.input, val, val)
			continue
		}
		if strVal.Value != tt.expected {
			t.Errorf("test input %q: expected %q, got %q", tt.input, tt.expected, strVal.Value)
		}
	}
}
