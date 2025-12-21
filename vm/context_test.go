package vm

import (
	"context"
	"testing"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/parser"
)

func TestBuiltinContextStorage(t *testing.T) {
	// Define two builtins:
	// setVal(k, v) -> sets context
	// getVal(k) -> gets context

	setVal := &object.Builtin{
		Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
			if len(args) != 2 {
				return &object.Error{Message: "wrong num args"}
			}
			key := args[0].(*object.String).Value
			// Store the object directly, or convert?
			// The interface Set(k string, v any) takes any.
			// Let's store the object's string representation for simplicity in this test,
			// or just the object itself.
			// Example usage: setVal("myKey", 123)
			val := args[1].(*object.Integer).Value
			ctx.Set(key, val)
			return object.NullObj
		},
	}

	getVal := &object.Builtin{
		Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
			if len(args) != 1 {
				return &object.Error{Message: "wrong num args"}
			}
			key := args[0].(*object.String).Value
			val, ok := ctx.Get(key)
			if !ok || val == nil {
				return object.NullObj
			}
			return &object.Integer{Value: val.(int64)}
		},
	}

	input := `
	setVal("test", 42);
	getVal("test");
	`

	// setup compiler
	c := compiler.New()

	// Define symbols (globals) manually so compiler knows about them
	// In the real world (REPL/Embedder), one would register these in symbol table and then set in VM
	setSym := c.SymbolTable().Define("setVal")
	getSym := c.SymbolTable().Define("getVal")

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	err := c.Compile(prog)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}

	vm := New(c.Bytecode())

	// Inject globals
	vm.SetGlobal(setSym.Index, setVal)
	vm.SetGlobal(getSym.Index, getVal)

	err = vm.Run(context.Background())
	if err != nil {
		t.Fatalf("vm run error: %s", err)
	}

	last := vm.LastPoppedStackElem()
	if last == nil {
		t.Fatal("last popped is nil")
	}

	integer, ok := last.(*object.Integer)
	if !ok {
		t.Fatalf("expected Integer, got %T", last)
	}

	if integer.Value != 42 {
		t.Errorf("expected 42, got %d", integer.Value)
	}
}
