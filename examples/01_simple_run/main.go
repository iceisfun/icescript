package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
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
	err = c.Compile(program)
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
