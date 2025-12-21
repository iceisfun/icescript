package vm

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/parser"
)

func TestBooleanTruthinessBug(t *testing.T) {
	// Define the builtin that flips from true to false
	flipFn := &object.Builtin{
		Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
			// Check if we have been called before
			_, called := ctx.Get("called")
			if !called {
				// First call: return true, mark as called
				ctx.Set("called", true)
				return &object.Boolean{Value: true}
			}
			// Subsequent calls: return false
			return &object.Boolean{Value: false}
		},
	}

	input := `
	// First call should be true
	if (test_isrunning()) {
		print("first: running")
	} else {
		print("first: not running")
	}

	// Second call should be false
	if (test_isrunning()) {
		print("second: running")
	} else {
		print("second: not running")
	}
	`

	// 1. Compile
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := compiler.New()
	// Register the symbol for the builtin
	idx := c.SymbolTable().Define("test_isrunning")
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}

	// 2. Setup VM
	vm := New(c.Bytecode())

	// Register the builtin function in the VM globals
	vm.SetGlobal(idx.Index, flipFn)

	// Capture output
	var out bytes.Buffer
	vm.SetOutput(&out)

	// 3. Run
	err = vm.Run(context.Background())
	if err != nil {
		t.Fatalf("vm run error: %s", err)
	}

	// 4. Verify output
	output := out.String()

	// Debug output
	t.Logf("Output:\n%s", output)

	// We expect:
	// first: running
	// second: not running

	if !strings.Contains(output, "first: running") {
		t.Error("First call did not take true branch")
	}
	if !strings.Contains(output, "second: not running") {
		t.Error("Second call did not take false branch (Bug: likely took true branch)")
	}
	if strings.Contains(output, "second: running") {
		t.Error("Second call INCORRECTLY took true branch (CONFIFRMED BUG)")
	}
}
