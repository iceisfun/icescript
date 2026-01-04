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
	// This tells the compiler that "Config" and "Callback" exist and assigns them indices.
	// We must use these indices to set the values in the VM later.
	configSym := c.SymbolTable().Define("Config")
	callbackSym := c.SymbolTable().Define("Callback")
	boolFunc := c.SymbolTable().Define("BoolFunc")

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
	// We use the indices we got from defining the symbols.

	// Inject "Config" (String)
	machine.SetGlobal(configSym.Index, &object.String{Value: "Production Mode"})

	// Inject "Callback" (Builtin Function)
	machine.SetGlobal(callbackSym.Index, &object.Builtin{
		Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
			// Arguments from Icescript
			a := args[0].(*object.Integer).Value
			b := args[1].(*object.Integer).Value
			fmt.Printf("[Host] Callback called with %d, %d\n", a, b)
			return &object.Integer{Value: a + b}
		},
	})

	// Test a boolean function
	machine.SetGlobal(boolFunc.Index, &object.Builtin{
		Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
			n := args[0].(*object.Integer).Value
			if n%2 == 0 {
				fmt.Println("[Host] BoolFunc: even number -- return true")
				return &object.Boolean{Value: true}
			}
			fmt.Println("[Host] BoolFunc: odd number -- return false")
			return &object.Boolean{Value: false}
		},
	})

	// 6. Run
	fmt.Println("Running script...")
	err = machine.Run(context.Background())
	if err != nil {
		log.Fatalf("VM error: %s", err)
	}
}
