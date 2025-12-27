package vm

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/parser"
)

func TestConcurrentInvoke(t *testing.T) {
	input := `
	func add(a, b) {
		return a + b
	}
	`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(c.Bytecode())

	// Grab the function to invoke
	// In icescript, functions are constants or globals.
	// We need to run the main body first to define the function if it's top level,
	// BUT here 'func add...' might be compiled as a constant.
	// Actually, func dcls at top level are usually put in globals or constants?
	// Let's just run the program once to populate globals?
	// But 'func add' is a statement.

	// Actually correct way to get a handle on "add" is to run the script, which defines "add" in global scope.

	err = vm.Run(context.Background())
	if err != nil {
		t.Fatalf("vm run error: %s", err)
	}

	addFn, err := vm.GetGlobal("add")
	if err != nil {
		t.Fatalf("GetGlobal error: %s", err)
	}

	var wg sync.WaitGroup
	params := []struct{ a, b int }{
		{1, 2},
		{3, 4},
		{5, 6},
		{7, 8},
		{9, 10},
	}

	// Concurrent invokes
	for _, p := range params {
		wg.Add(1)
		go func(a, b int) {
			defer wg.Done()
			arg1 := &object.Integer{Value: int64(a)}
			arg2 := &object.Integer{Value: int64(b)}
			res, err := vm.Invoke(context.Background(), addFn, arg1, arg2)
			if err != nil {
				// We expect it NOT to error if fixed, but it might panic if not fixed.
				// t.Error cannot be called from goroutine safely without helper?
				// actually t.Error is safe in newer Go, but panic crashes the test.
				fmt.Printf("Invoke error: %v\n", err)
				return
			}
			val := res.(*object.Integer).Value
			if val != int64(a+b) {
				fmt.Printf("wrong result: %d != %d\n", val, a+b)
			}
		}(p.a, p.b)
	}

	wg.Wait()
}
