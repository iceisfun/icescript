package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/parser"
	"github.com/iceisfun/icescript/vm"
)

func main() {
	input := `
	var x = 10
	var y = 20
	print("Hello directly from icescript!")
	print("x + y =", x + y)
	`

	// 1. Lexing
	l := lexer.New(input)

	// 2. Parsing
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		log.Fatalf("Parse errors: %v", p.Errors())
	}

	// 3. Compiling
	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		log.Fatalf("Compiler error: %s", err)
	}

	// 4. Execution (VM)
	machine := vm.New(c.Bytecode())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Starting execution...")
	start := time.Now()
	err = machine.Run(ctx)
	if err != nil {
		log.Fatalf("VM error: %s", err)
	}
	fmt.Printf("Execution finished in %s\n", time.Since(start))
}
