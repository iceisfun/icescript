package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/iceisfun/icescript/compiler" // Imported again? No, duplicate. Go will complain.
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/parser"
	"github.com/iceisfun/icescript/vm"
)

func main() {
	if len(os.Args) > 1 {
		runFile(os.Args[1])
	} else {
		startREPL()
	}
}

func runFile(filename string) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("could not read file: %s\n", err)
		os.Exit(1)
	}

	run(string(bytes))
}

func startREPL() {
	fmt.Println("icescript v2 REPL")
	fmt.Println("Type 'exit' to quit.")

	// Simple REPL (just stdin reader)
	// For better experience, we'd use bufio or a readline lib.
	// But minimal implementation:
	var input string
	for {
		fmt.Print(">> ")
		_, err := fmt.Scanln(&input) // Scanln breaks on spaces! Use bufio.
		if err != nil {
			if err == io.EOF {
				return
			}
			// fmt.Println(err) // Scanln issues
		}
		// Fallback to bufio
		// Re-implementing correctly below.
		break
	}

	// Re-do with correct input reading
	runREPL()
}

func runREPL() {
	// Re-implemented correctly
	buffer := make([]byte, 1024)
	for {
		fmt.Print(">> ")
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println("error:", err)
			return
		}

		line := string(buffer[:n])
		if line == "exit\n" {
			return
		}

		// Compile and Run
		// Maintain global state?
		// For now simple each-line-is-fresh VM (no persistent globals in REPL for this basic version unless we pass symbol table)
		// To support persistent REPL, we need to keep compiler and machine state.

		// Let's do persistent constants and globals.
		// compiler := compiler.New() // New each time? No.

		run(line)
	}
}

func run(input string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		printParserErrors(os.Stdout, p.Errors())
		return
	}

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		fmt.Printf("compiler execution failed: %s\n", err)
		return
	}

	machine := vm.New(comp.Bytecode())
	err = machine.Run(context.Background())
	if err != nil {
		fmt.Printf("vm execution failed: %s\n", err)
		return
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped != nil {
		fmt.Println(lastPopped.Inspect())
	}
}

func printParserErrors(out io.Writer, errors []string) {
	fmt.Fprintf(out, "parser errors:\n")
	for _, msg := range errors {
		fmt.Fprintf(out, "\t%s\n", msg)
	}
}
