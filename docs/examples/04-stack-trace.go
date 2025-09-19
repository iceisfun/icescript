//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/iceisfun/icescript/icescript"
)

func main() {
	script := `
func boom() {
  var nums = [1, 0, 2]
  var total = 0
  for n in nums {
    total = total + 10 / n
  }
  return total
}

func fn1() {
  return fn2()
}

func fn2() {
  return fn3()
}

func fn3() {
  return boom()
}
`

	vm := icescript.NewVM()
	if err := vm.Compile(script); err != nil {
		panic(err)
	}

	if rval, err := vm.Invoke("fn1"); err != nil {
		fmt.Println("runtime error:")
		fmt.Println(err)
	} else {
		fmt.Println("Return value:", rval.String())
	}
}
