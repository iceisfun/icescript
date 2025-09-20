package icescript

import "testing"

// TestObjectNestedMutation ensures that assignments to nested
// object fields in script update the underlying Go value.
func TestObjectNestedMutation(t *testing.T) {
	vm := NewVM()

	const src = `
func bump() {
  Player.Life = Player.Life + 1
  Player.Mana = Player.Mana + 2
  return { life: Player.Life, mana: Player.Mana }
}
`

	// Compile the script
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile bump script: %v", err)
	}

	// Provide a Player object into the VM
	type Player struct {
		Name string
		Life int
		Mana int
		X, Y float64
	}
	p := Player{Name: "Hero", Life: 10, Mana: 20}
	vm.SetGlobal("Player", MustVFromGo(p))

	// Invoke bump()
	out, err := vm.Invoke("bump")
	if err != nil {
		t.Fatalf("invoke bump: %v", err)
	}

	// Verify the return object from script
	obj := out.Obj
	if obj["life"].AsInt() != 11 {
		t.Fatalf("expected life=11, got %d", obj["life"].AsInt())
	}
	if obj["mana"].AsInt() != 22 {
		t.Fatalf("expected mana=22, got %d", obj["mana"].AsInt())
	}

	// Also pull Player back into Go to ensure mutation persisted
	var updated Player
	if err := vm.GetGlobal("Player").ToGo(&updated); err != nil {
		t.Fatalf("GetGlobal(Player).ToGo: %v", err)
	}
	if updated.Life != 11 || updated.Mana != 22 {
		t.Fatalf("Go Player not updated correctly: %+v", updated)
	}
}
