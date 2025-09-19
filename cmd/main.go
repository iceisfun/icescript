package main

import (
	"fmt"
	"lex/icescript"
	"math"
)

func main() {
	src := `
func add(a int, b int) int {
  print("hello world", "we add", a, b, f, distance(0, 0, 1, 1))
  return a + b + distance(0, 0, 1, 1)
}
`

	vm := icescript.NewVM()

	// host function
	vm.RegisterHostFunc("print", func(_ *icescript.VM, argv []icescript.Value) (icescript.Value, error) {
		for i, v := range argv {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(v.String())
		}
		fmt.Println()
		return icescript.VNull(), nil
	})

	vm.RegisterHostFunc("distance", func(_ *icescript.VM, argv []icescript.Value) (icescript.Value, error) {
		if len(argv) != 4 {
			return icescript.VNull(), fmt.Errorf("distance requires 4 arguments")
		}
		x1, y1 := argv[0].AsFloat(), argv[1].AsFloat()
		x2, y2 := argv[2].AsFloat(), argv[3].AsFloat()
		dx, dy := x2-x1, y2-y1
		return icescript.VFloat(math.Hypot(dx, dy)), nil
	})

	// globals
	vm.SetGlobal("PI", icescript.VFloat(3.14159))
	vm.SetGlobal("f", icescript.VFloat(42.0))

	if err := vm.Compile(src); err != nil {
		panic(err)
	}

	out, err := vm.Invoke("add", icescript.VFloat(0), icescript.VFloat(0))
	if err != nil {
		panic(err)
	}
	fmt.Println("result:", out.AsFloat())

}
