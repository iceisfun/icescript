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
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <script.ice>")
	}
	filename := os.Args[1]

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	input := string(data)

	// 2. Setup SymbolTable
	// We must define builtins because custom symbol table starts empty
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	// Define globals to match VM state we will provide
	// Order matters for index if we used auto-index, but Define returns the symbol.
	versionSym := symbolTable.Define("version")
	configSym := symbolTable.Define("config")

	// 3. Compile
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		log.Fatalf("Parse errors: %v", p.Errors())
	}

	// NewWithState accepts just constants if we want, or symbol table
	// compiler.NewWithState(symbolTable, constants)
	c := compiler.NewWithState(symbolTable, []object.Object{})
	err = c.Compile(program)
	if err != nil {
		log.Fatalf("Compiler error: %s", err)
	}

	// 4. Prepare VM Globals
	// We need a slice large enough
	globals := make([]object.Object, vm.GlobalSize)

	// Populate inputs
	globals[versionSym.Index] = &object.Integer{Value: 1}

	// Create a Hash for config
	confMap := make(map[object.HashKey]object.HashPair)

	keyEnv := &object.String{Value: "env"}
	valEnv := &object.String{Value: "production"}
	confMap[keyEnv.HashKey()] = object.HashPair{Key: keyEnv, Value: valEnv}

	globals[configSym.Index] = &object.Hash{Pairs: confMap}

	fmt.Println("Go: Global 'version' set to 1")
	fmt.Println("Go: Global 'config' set to", globals[configSym.Index].Inspect())

	// 5. Run VM
	machine := vm.NewWithGlobalsStore(c.Bytecode(), globals)
	err = machine.Run(context.Background())
	if err != nil {
		log.Fatalf("VM error: %s", err)
	}

	// 6. Retrieve results
	finalVersion := globals[versionSym.Index]
	finalConfig := globals[configSym.Index]

	fmt.Printf("\nGo: Final Version: %s\n", finalVersion.Inspect())
	fmt.Printf("Go: Final Config: %s\n", finalConfig.Inspect())
}
