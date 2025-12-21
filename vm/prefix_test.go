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

func TestPrintPrefix(t *testing.T) {
	input := `print("hello world")`

	// 1. Compile
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

	// 2. Setup VM
	vm := New(c.Bytecode())

	// Capture output
	var out bytes.Buffer
	vm.SetOutput(&out)

	// Set Prefix
	prefix := "[LOG]"
	vm.SetPrintPrefix(prefix)

	// 3. Run
	err = vm.Run(context.Background())
	if err != nil {
		t.Fatalf("vm run error: %s", err)
	}

	// 4. Verify
	expected := "[LOG] hello world\n"
	if !strings.Contains(out.String(), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, out.String())
	}
}
