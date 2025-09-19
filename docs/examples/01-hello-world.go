//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/iceisfun/icescript/icescript"
)

func main() {
	const src = `
func main() null {
  print("hello from icescript")
  return null
}
`

	vm := icescript.NewVM()
	vm.RegisterHostFunc("print", func(_ *icescript.VM, args []icescript.Value) (icescript.Value, error) {
		for i, v := range args {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(v.String())
		}
		fmt.Println()
		return icescript.VNull(), nil
	})

	if err := vm.Compile(src); err != nil {
		panic(err)
	}

	if _, err := vm.Invoke("main"); err != nil {
		panic(err)
	}
}
