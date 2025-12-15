package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/iceisfun/icescript/compiler" // Imported again? No, duplicate. Go will complain.
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/object"
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

	scanner := bufio.NewScanner(os.Stdin)

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalSize)
	symbolTable := compiler.NewSymbolTable()

	var accumulatedInput strings.Builder
	var openBraces int

	for {
		if openBraces > 0 {
			fmt.Print("... ")
		} else {
			fmt.Print(">> ")
		}

		if !scanner.Scan() {
			return
		}

		line := scanner.Text()
		if line == "exit" {
			return
		}

		accumulatedInput.WriteString(line)
		accumulatedInput.WriteString("\n")

		openBraces += strings.Count(line, "{") - strings.Count(line, "}")

		if openBraces > 0 {
			continue
		}

		input := accumulatedInput.String()
		accumulatedInput.Reset()
		openBraces = 0 // Reset just in case

		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			printParserErrors(os.Stdout, p.Errors())
			continue
		}

		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compiler execution failed: %s\n", err)
			continue
		}

		// Update constants for next run
		code := comp.Bytecode()
		constants = code.Constants

		machine := vm.NewWithGlobalsStore(code, globals)
		err = machine.Run(context.Background())
		if err != nil {
			fmt.Printf("vm execution failed: %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElem()
		if lastPopped != nil {
			fmt.Println(lastPopped.Inspect())
		}
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
