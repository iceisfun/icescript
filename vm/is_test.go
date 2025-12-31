package vm

import (
	"context"
	"testing"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/object"
)

func TestIsOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// Integer
		{"1 is int", true},
		{"1 is integer", true},
		{"1 is float", false},
		{"1 is string", false},

		// Float
		{"1.5 is float", true},
		{"1.5 is int", false},

		// Boolean
		{"true is bool", true},
		{"false is boolean", true},
		{"true is int", false},

		// String
		{`"hello" is str`, true},
		{`"hello" is string`, true},
		{`"hello" is array`, false},

		// Array
		{"[1, 2] is array", true},
		{"[1, 2] is tuple", false},

		// Null
		{"null is null", true},
		{"1 is null", false},

		// Tuple
		{"testMultiReturn(1, 2) is tuple", true},
		{"testMultiReturn(1, 2) is int", false},

		// User
		// Need a way to create USER_OBJ. Builtins?
		// Maybe just trust logic for now, or add a fake USER_OBJ in test if possible?
		// We don't have easy way to create UserObj from script.

		// Precedence
		{"1 is int && 1 == 1", true},
		{"1 is int && 1 == 2", false},
		{"1 is float && 1 == 1", false},
		{"1 == 1 && 1 is int", true},
	}

	for _, tt := range tests {
		val := testEval(t, tt.input)
		boolVal, ok := val.(*object.Boolean)
		if !ok {
			t.Errorf("test input %q: expected boolean, got %T (%+v)", tt.input, val, val)
			continue
		}
		if boolVal.Value != tt.expected.(bool) {
			t.Errorf("test input %q: expected %t, got %t", tt.input, tt.expected, boolVal.Value)
		}
	}
}

func testEval(t *testing.T, input string) object.Object {
	program := parse(input)
	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(comp.Bytecode())
	err = vm.Run(context.Background())
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	return vm.LastPoppedStackElem()
}
