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

	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors:\n")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	c := compiler.New()
	err = c.Compile(program)
	if err != nil {
		fmt.Printf("Compiler error: %s\n", err)
		os.Exit(1)
	}

	machine := vm.New(c.Bytecode())
	err = machine.Run(context.Background())
	if err != nil {
		fmt.Printf("VM execution failed:\n%s\n", err)
		// We expect failure here, so exit 0 if it's the expected panic
		// But in real world app, failure is failure.
		// For demo, we just print it.
		os.Exit(1)
	} else {
		// If it succeeded, we failed the test script?
		// But maybe we want to see output.
		fmt.Println("Execution finished successfully (unexpectedly).")
		result := machine.LastPoppedStackElem()
		if result != nil {
			fmt.Printf("Last value: %s\n", result.Inspect())
		}
	}
}
