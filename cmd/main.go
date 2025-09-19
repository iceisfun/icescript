package main

import (
	"fmt"
	"math"

	"github.com/iceisfun/icescript/icescript"
)

type Player struct {
	Name string
	Life int
	Mana int
	X    float64
	Y    float64
}

func main() {
	src := `
func demo() {
  print("Player")
  print("Player.Name:", Player.Name, "Life:", Player.Life, "Mana:", Player.Mana)
  Player.X = Player.X + 1
  Player.Y = Player.Y + 1
  print("Player position:", Player.X, Player.Y)

  var p = { x: 3, y: 4 }
  print("p:", p, "len:", distance(0, 0, p.x, p.y))

  var nums = [1, 2, 3, 4, 5]
  var total = 0
  for n in nums {
    print("n:", n)
    total = total + n
  }
  print("sum:", total)

  // index access
  print("first:", nums[0], "last:", nums[4])

  return total
}
`
	vm := icescript.NewVM()

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
		return icescript.VFloat(math.Hypot(x2-x1, y2-y1)), nil
	})

	p := Player{Name: "Hero", Life: 100, Mana: 50, X: 23, Y: -57}
	vm.SetGlobal("Player", icescript.MustVFromGo(p))

	if err := vm.Compile(src); err != nil {
		panic(err)
	}

	for range 5 {
		out, err := vm.Invoke("demo")
		if err != nil {
			panic(err)
		}
		fmt.Println("demo() ->", out.AsInt())
	}

	if err := vm.GetGlobal("Player").ToGo(&p); err != nil {
		panic(err)
	}
	fmt.Println("Go Player:", p)
}
