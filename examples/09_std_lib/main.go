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

	l := lexer.New(string(input))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Print("Parser errors:\n")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	c := compiler.New()
	err = c.Compile(program)
	if err != nil {
		fmt.Printf("Compiler error:\n%s\n", err)
		os.Exit(1)
	}

	machine := vm.New(c.Bytecode())
	err = machine.Run(context.Background())
	if err != nil {
		fmt.Printf("VM execution failed:\n%s\n", err)
		os.Exit(1)
	}
}
