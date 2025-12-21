package vm

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/parser"
)

func TestExtendedTruthiness(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`if (0) { print("truthy") } else { print("falsy") }`, "falsy"},
		{`if (1) { print("truthy") } else { print("falsy") }`, "truthy"},
		{`if ("") { print("truthy") } else { print("falsy") }`, "falsy"},
		{`if ("hello") { print("truthy") } else { print("falsy") }`, "truthy"},
		{`if (0.0) { print("truthy") } else { print("falsy") }`, "falsy"},
		{`if (0.1) { print("truthy") } else { print("falsy") }`, "truthy"},
		{`if (null) { print("truthy") } else { print("falsy") }`, "falsy"},
		{`if (true) { print("truthy") } else { print("falsy") }`, "truthy"},
		{`if (false) { print("truthy") } else { print("falsy") }`, "falsy"},
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
		var out bytes.Buffer
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
