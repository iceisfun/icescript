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
  var obj = {
    sqrt: sqrt(16),
    dist: distance(0, 0, 6, 8),
    sin0: sin(0),
    cos0: cos(0),
    atan1: atan(1),
    abs: abs(-12),
    lenStr: len("hello"),
    lower: lower("WORLD"),
    contains: contains("or", "world"),
  }
  sleep(0)
  return obj
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
