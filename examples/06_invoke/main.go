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
	// 1. Load and compile script
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <script.ice>")
	}
	filename := os.Args[1]

	input, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	l := lexer.New(string(input))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, msg := range p.Errors() {
			fmt.Println("Parse Error:", msg)
		}
		panic("Parse errors encountered")
	}

	c := compiler.New()
	err = c.Compile(program)
	if err != nil {
		panic(err)
	}

	bytecode := c.Bytecode()
	machine := vm.New(bytecode)

	// 2. Initialize state (run the script once)
	ctx := context.Background()
	if err := machine.Run(ctx); err != nil {
		panic(err)
	}

	// 3. Get handles to functions
	onTick, err := machine.GetGlobal("onTick")
	if err != nil {
		panic(err)
	}

	add, err := machine.GetGlobal("add")
	if err != nil {
		panic(err)
	}

	slow, err := machine.GetGlobal("slow")
	if err != nil {
		panic(err)
	}

	// 4. Invoke repeatedly
	fmt.Println("--- Invoking onTick ---")
	for i := 0; i < 3; i++ {
		result, err := machine.Invoke(ctx, onTick, &object.Integer{Value: int64(i * 10)})
		if err != nil {
			panic(err)
		}
		fmt.Println("Create Result:", result.Inspect())
	}

	fmt.Println("--- Invoking add ---")
	res, err := machine.Invoke(ctx, add, &object.Integer{Value: 5}, &object.Integer{Value: 7})
	if err != nil {
		panic(err)
	}
	fmt.Println("5 + 7 =", res.Inspect())

	// Test context cancellation
	fmt.Println("--- Testing Cancellation ---")
	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	_, err = machine.Invoke(ctxCancel, slow)
	if err != nil {
		fmt.Println("Got expected error:", err)
	}
}
