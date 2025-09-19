//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/iceisfun/icescript/icescript"
)

func main() {
	const src = `
func add(a, b int) {
  return a + b
}

func twice(x int) {
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
}
