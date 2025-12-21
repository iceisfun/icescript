package vm

import (
	"context"
	"strings"
	"testing"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/parser"
)

func TestStrictTruthinessError(t *testing.T) {
	// Array in condition should error now
	input := `if ([]) { print("should not print") }`

	l := lexer.New(input)
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

	err = vm.Run(context.Background())
	if err == nil {
		t.Fatal("expected runtime error for array in condition, got nil")
	}

	expectedError := "condition must be boolean, got ARRAY"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("expected error containing %q, got %q", expectedError, err.Error())
	}
}
