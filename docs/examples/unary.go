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
  var x = 5
  var flag = true
  x++
  x--
  return { neg: -x, pos: +x, not: !flag }
}
`

	vm := icescript.NewVM()
	if err := vm.Compile(src); err != nil {
		panic(err)
	}

	out, err := vm.Invoke("demo")
	if err != nil {
		panic(err)
	}
	fmt.Println("demo() ->", out.String())
}
