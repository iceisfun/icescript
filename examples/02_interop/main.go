package main

import (
	"context"
	"fmt"
	"log"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/parser"
	"github.com/iceisfun/icescript/vm"
)

func main() {
	input := `
	// "Config" and "Callback" are NOT declared here with var.
	// They are injected by the host.

	print("Config is:", Config)
	
	var result = Callback(1, 2)
	print("Callback returned:", result)
	`

	// 1. Setup Compiler
	c := compiler.New()

	// 2. Pre-define globals in the symbol table
	// This tells the compiler that "Config" and "Callback" exist and assigns them indices.
	// We must use these indices to set the values in the VM later.
	configSym := c.SymbolTable().Define("Config")
	callbackSym := c.SymbolTable().Define("Callback")

	// 3. Compile
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		log.Fatalf("Parse errors: %v", p.Errors())
	}

	err := c.Compile(program)
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
		Fn: func(args ...object.Object) object.Object {
			// Arguments from Icescript
			a := args[0].(*object.Integer).Value
			b := args[1].(*object.Integer).Value
			fmt.Printf("[Host] Callback called with %d, %d\n", a, b)
			return &object.Integer{Value: a + b}
		},
	})

	// 6. Run
	fmt.Println("Running script...")
	err = machine.Run(context.Background())
	if err != nil {
		log.Fatalf("VM error: %s", err)
	}
}
