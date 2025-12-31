package vm

import (
	"context"
	"strings"
	"testing"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/parser"
	"github.com/iceisfun/icescript/token"
)

func TestRuntimeErrorMetadata(t *testing.T) {
	input := `
	var f = func() {
		return 1 + "wrong" // Type mismatch on line 3
	}
	f()
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.StructuredErrors()) > 0 {
		t.Fatalf("parser errors: %v", p.StructuredErrors())
	}

	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(c.Bytecode())
	err = vm.Run(context.Background())

	if err == nil {
		t.Fatal("expected runtime error, got nil")
	}

	scriptErr, ok := err.(*token.ScriptError)
	if !ok {
		t.Fatalf("expected *token.ScriptError, got %T: %v", err, err)
	}

	if scriptErr.Kind != token.ErrorKindRuntime {
		t.Errorf("expected Kind=Runtime, got %v", scriptErr.Kind)
	}

	expectedLine := 3
	if scriptErr.Line != expectedLine {
		t.Errorf("expected line %d, got %d", expectedLine, scriptErr.Line)
	}

	// Check stack trace contains our function name (which is anonymous, but let's check basic structure)
	// The function is assigned to 'f', but the literal itself is anonymous initially unless name binding happens.
	// In the assignment `var f = func...`, the compiled function usually gets name "f" if the compiler is smart about LetStatement,
	// or it remains empty.
	// Let's check if my compiler update handles `var f = ...` naming?
	// The provided compiler.go snippet in context shows:
	// case *ast.FunctionLiteral: ...
	// It doesn't seem to infer name from LetStatement context automatically unless explicitly named `func f()`.
	// However, `stackTrace` should contain "anonymous" or empty.

	if len(scriptErr.StackTrace) == 0 {
		t.Error("expected stack trace, got empty")
	} else {
		// Just ensure it's formatted reasonably
		topFrame := scriptErr.StackTrace[0]
		if !strings.Contains(topFrame, "line 3") {
			t.Errorf("stack frame missing line number: %s", topFrame)
		}
	}
}

func TestParserErrorMetadata(t *testing.T) {
	input := `
	var x = 1
	var y = ; // Syntax error line 3
	`
	l := lexer.New(input)
	p := parser.New(l)
	p.ParseProgram()

	errs := p.StructuredErrors()
	if len(errs) == 0 {
		t.Fatal("expected parser errors")
	}

	firstErr := errs[0]
	if firstErr.Line != 3 {
		t.Errorf("expected error on line 3, got %d", firstErr.Line)
	}
}
