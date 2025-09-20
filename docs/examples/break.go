//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/iceisfun/icescript/icescript"
)

func main() {
	const src = `
func demo() {
  var nums = [1, 2, 3, 4]
  var total = 0
  for n in nums {
    if (n == 3) {
      break
    }
    total = total + n
  }
  return total
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

	out, err := vm.Invoke("demo")
	if err != nil {
		panic(err)
	}
	fmt.Println("demo() ->", out.AsInt())
}
