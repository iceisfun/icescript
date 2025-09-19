package icescript

import (
	"testing"
)

func mustCompile(t *testing.T, src string, setup func(*VM)) *VM {
	t.Helper()
	vm := NewVM()
	if setup != nil {
		setup(vm)
	}
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile error: %v", err)
	}
	return vm
}

func TestVMInvokeReturnValue(t *testing.T) {
	src := `
func sum(a, b) {
  return a + b
}
`
	vm := mustCompile(t, src, nil)
	out, err := vm.Invoke("sum", VInt(2), VInt(3))
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if got := out.AsInt(); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestVMHostFunctionBinding(t *testing.T) {
	src := `
func call() {
  return add(2, 3)
}
`
	calls := 0
	vm := mustCompile(t, src, func(vm *VM) {
		vm.RegisterHostFunc("add", func(_ *VM, args []Value) (Value, error) {
			calls++
			if len(args) != 2 {
				t.Fatalf("expected 2 args, got %d", len(args))
			}
			return VInt(args[0].AsInt() + args[1].AsInt()), nil
		})
	})

	out, err := vm.Invoke("call")
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if out.AsInt() != 5 {
		t.Fatalf("expected 5, got %d", out.AsInt())
	}
	if calls != 1 {
		t.Fatalf("expected host function to be called once, got %d", calls)
	}
}

func TestVMIfElseMutatesGlobal(t *testing.T) {
	src := `
func decide(damage) {
  if (damage > 0) {
    Player.Life = Player.Life - damage
  } else {
    Player.Life = Player.Life + 1
  }
  return null
}
`

	type player struct {
		Name string
		Life int
	}

	vm := mustCompile(t, src, func(vm *VM) {
		vm.SetGlobal("Player", MustVFromGo(player{Name: "Hero", Life: 10}))
	})

	if _, err := vm.Invoke("decide", VInt(3)); err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	var updated player
	if err := vm.GetGlobal("Player").ToGo(&updated); err != nil {
		t.Fatalf("ToGo failed: %v", err)
	}
	if updated.Life != 7 {
		t.Fatalf("expected life 7 after hit, got %d", updated.Life)
	}

	if _, err := vm.Invoke("decide", VInt(0)); err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if err := vm.GetGlobal("Player").ToGo(&updated); err != nil {
		t.Fatalf("ToGo failed: %v", err)
	}
	if updated.Life != 8 {
		t.Fatalf("expected life 8 after heal, got %d", updated.Life)
	}
}
