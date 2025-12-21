package vm

import (
	"context"
	"strings"
	"testing"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/parser"
)

func TestUnaryMinus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"print(-1)", "-1"},
		{"print(-10)", "-10"},
		{"print(-1.5)", "-1.500000"},
		{"print(-0.01)", "-0.010000"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			t.Fatalf("parse errors: %v", p.Errors())
		}

		c := compiler.New()
		err := c.Compile(program)
		if err != nil {
			t.Fatalf("compile error: %s", err)
		}

		vm := New(c.Bytecode())
		var out strings.Builder
		vm.SetOutput(&out)

		err = vm.Run(context.Background())
		if err != nil {
			t.Fatalf("vm run error: %s", err)
		}

		if !strings.Contains(out.String(), tt.expected) {
			t.Errorf("input: %s\nexpected: %s\ngot: %s", tt.input, tt.expected, out.String())
		}
	}
}
