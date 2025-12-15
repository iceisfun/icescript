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

	start := time.Now()

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

	compileTime := time.Since(start)

	machine := vm.New(c.Bytecode())

	startExec := time.Now()
	err = machine.Run(context.Background())
	if err != nil {
		log.Fatalf("VM error: %s", err)
	}
	execTime := time.Since(startExec)

	fmt.Printf("\n--- Stats ---\nCompile time: %s\nExecution time: %s\n", compileTime, execTime)
}
