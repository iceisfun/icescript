package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		log.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.New()
	err = c.Compile(program)
	if err != nil {
		log.Fatalf("Compiler error: %s", err)
	}

	machine := vm.New(c.Bytecode())

	err = machine.Run(context.Background())
	if err != nil {
		log.Fatalf("VM error: %s", err)
	}

	fmt.Println("\nExecution finished.")
}
