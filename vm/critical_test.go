package vm

import (
	"context"
	"strings"
	"testing"

	"github.com/iceisfun/icescript/compiler"
)

func TestCriticalRuntimeErrors(t *testing.T) {
	tests := []struct {
		input       string
		errContains string
	}{
		{
			`len(1, 2);`,
			"wrong number of arguments",
		},
		{
			`len();`,
			"wrong number of arguments",
		},
		{
			`push();`,
			"wrong number of arguments",
		},
		{
			`push(1, 2);`,
			"argument to `push` must be ARRAY",
		},
		{
			`testMultiReturn(1);`,
			"wrong number of arguments",
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
