//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/iceisfun/icescript/icescript"
)

func main() {
	const src = `
func add(a, b) {
  return a + b
}

func twice(x foobar) {
  return add(x, x)
}
`

	parser := icescript.NewParser(icescript.New(src))
	program, errs := parser.ParseProgram()
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println("parse error:", err)
		}
		return
	}

	fmt.Printf("program has %d functions\n", len(program.Funcs))
	for name, fn := range program.Funcs {
		fmt.Printf("- %s (%d params) defined at %d:%d\n", name, len(fn.Params), fn.Start.Line, fn.Start.Column)
	}

	// run the program calling the 'twice' function
	vm := icescript.NewVM()
	if err := vm.Compile(src); err != nil {
		fmt.Println("load program error:", err)
		return
	}

	out, err := vm.Invoke("twice", icescript.VInt(21))
	if err != nil {
		fmt.Println("invoke error:", err)
		return
	}
	fmt.Println("twice(21) =", out.String())
}
