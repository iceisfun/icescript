//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"math"

	"github.com/iceisfun/icescript/icescript"
)

func main() {
	script := `
func main() float {
  return distance(0, 0, 3, 4)
}
`

	vm := icescript.NewVM()
	vm.RegisterHostFunc("distance", func(_ *icescript.VM, args []icescript.Value) (icescript.Value, error) {
		if len(args) != 4 {
			return icescript.VNull(), fmt.Errorf("distance expects 4 args")
		}
		ax, ay := args[0].AsFloat(), args[1].AsFloat()
		bx, by := args[2].AsFloat(), args[3].AsFloat()
		return icescript.VFloat(math.Hypot(bx-ax, by-ay)), nil
	})

	if err := vm.Compile(script); err != nil {
		panic(err)
	}

	out, err := vm.Invoke("main")
	if err != nil {
		panic(err)
	}

	fmt.Printf("distance: %.1f\n", out.AsFloat())
}
