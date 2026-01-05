package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		log.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.New()
	boolFunc := c.SymbolTable().Define("BoolFunc")

	err = c.Compile(program)
	if err != nil {
		log.Fatalf("Compiler error: %s", err)
	}

	machine := vm.New(c.Bytecode())
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Running comprehensive tests...")
	start := time.Now()
	err = machine.Run(ctx)
	if err != nil {
		log.Fatalf("VM error: %s", err)
	}
	fmt.Printf("All tests completed in %s\n", time.Since(start))
}
