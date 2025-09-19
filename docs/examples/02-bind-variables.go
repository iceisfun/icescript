//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/iceisfun/icescript/icescript"
)

type Player struct {
	Name string
	Life int
	Mana int
	X, Y float64
}

func main() {
	script := `
func move(dx, dy) null {
  Player.X = Player.X + dx
  Player.Y = Player.Y + dy
  return null
}
`

	vm := icescript.NewVM()
	vm.SetGlobal("Player", icescript.MustVFromGo(Player{Name: "Hero", Life: 100, Mana: 50}))

	if err := vm.Compile(script); err != nil {
		panic(err)
	}

	if _, err := vm.Invoke("move", icescript.VFloat(3), icescript.VFloat(-2)); err != nil {
		panic(err)
	}

	var out Player
	if err := vm.GetGlobal("Player").ToGo(&out); err != nil {
		panic(err)
	}

	fmt.Printf("updated player: %+v\n", out)
}
