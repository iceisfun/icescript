package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/parser"
	"github.com/iceisfun/icescript/vm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <script.ice>")
	}
	filename := os.Args[1]

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	input := string(data)

	// 1. Setup Compiler
	c := compiler.New()

	// 2. Pre-define globals in the symbol table
	configSym := c.SymbolTable().Define("Config")
	newGotypeSym := c.SymbolTable().Define("NewGotype")
	incStateSym := c.SymbolTable().Define("IncState")
	getStateSym := c.SymbolTable().Define("GetState")

	// 3. Compile
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		log.Fatalf("Parse errors: %v", p.Errors())
	}

	err = c.Compile(program)
	if err != nil {
		log.Fatalf("Compiler error: %s", err)
	}

	// 4. Setup VM
	machine := vm.New(c.Bytecode())

	// 5. Inject Values into VM Globals

	// Config
	machine.SetGlobal(configSym.Index, &object.String{Value: "Production Mode"})

	// NewGotype(a, b) -> *Gotype
	machine.SetGlobal(newGotypeSym.Index, &object.Builtin{
		Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
			// Simply returns a new Gotype
			fmt.Println("[Host] NewGotype called")
			return &object.User{Value: NewGotype()}
		},
	})

	// IncState(obj)
	machine.SetGlobal(incStateSym.Index, &object.Builtin{
		Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
			if len(args) < 1 {
				return &object.Error{Message: "IncState requires 1 argument"}
			}
			// Safe casting
			u, ok := args[0].(*object.User)
			if !ok {
				return &object.Error{Message: "Argument must be a User object"}
			}
			g, ok := u.Value.(*Gotype)
			if !ok {
				return &object.Error{Message: "User object is not a Gotype"}
			}

			g.internalState++
			fmt.Printf("[Host] Incremented state to %d\n", g.internalState)
			return object.NativeBoolToBooleanObject(true)
		},
	})

	// GetState(obj) -> int
	machine.SetGlobal(getStateSym.Index, &object.Builtin{
		Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
			if len(args) < 1 {
				return &object.Error{Message: "GetState requires 1 argument"}
			}
			u, ok := args[0].(*object.User)
			if !ok {
				return &object.Error{Message: "Argument must be a User object"}
			}
			g, ok := u.Value.(*Gotype)
			if !ok {
				return &object.Error{Message: "User object is not a Gotype"}
			}

			return &object.Integer{Value: int64(g.internalState)}
		},
	})

	// 6. Run
	fmt.Println("Running script...")
	err = machine.Run(context.Background())
	if err != nil {
		log.Fatalf("VM error: %s", err)
	}
}
