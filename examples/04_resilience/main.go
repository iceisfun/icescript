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
	print("Starting infinite loop...")
	for {
		// spin
	}
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		log.Fatalf("Parse errors: %v", p.Errors())
	}

	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		log.Fatalf("Compiler error: %s", err)
	}

	machine := vm.New(c.Bytecode())

	// Timeout of 200ms
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	fmt.Println("Run with 200ms timeout...")
	start := time.Now()
	err = machine.Run(ctx)

	if err != nil {
		if err == context.DeadlineExceeded {
			fmt.Println("Success! Execution timed out as expected.")
		} else {
			fmt.Printf("Execution failed with %v\n", err)
		}
	} else {
		fmt.Println("Execution finished (unexpectedly!)")
	}
	fmt.Printf("Waited: %s\n", time.Since(start))
}
