package vm

import (
	"context"
	"strings"
	"testing"

	"github.com/iceisfun/icescript/compiler"
)

func TestMultiReturnDestructuring(t *testing.T) {
	tests := []vmTestCase{
		{
			`
			var a, b = testMultiReturn(1, 2);
			a;
			`,
			1,
		},
		{
			`
			var a, b = testMultiReturn(1, 2);
			b;
			`,
			2,
		},
		{
			`
			var a, b = testMultiReturn(1, 2);
			a + b;
			`,
			3,
		},
	}
	runVmTests(t, tests)
}

func TestMultiReturnBackwardCompat(t *testing.T) {
	tests := []vmTestCase{
		{
			`
			var a = testMultiReturn(10, 20);
			a; // Should be 10 (first element)
			`,
			10,
		},
		{
			`
			var a = testMultiReturn(10, 20);
			a + 5; // Should treat tuple as 10
			`,
			15,
		},
	}
	runVmTests(t, tests)
}

func TestMultiReturnRuntimeErrors(t *testing.T) {
	tests := []struct {
		input       string
		errContains string
	}{
		{
			`
			var a, b, c = testMultiReturn(1, 2);
			`,
			"not enough values to unpack",
		},
		{
			`
			var a, b = 1; // scalar unpack
			`,
			"cannot destructure non-tuple",
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run(context.Background())
		if err == nil {
			t.Errorf("expected error containing %q, got nil", tt.errContains)
		} else if !strings.Contains(err.Error(), tt.errContains) {
			t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
		}
	}
}
