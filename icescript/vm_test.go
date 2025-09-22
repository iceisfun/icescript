package icescript

import (
	"strings"
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

func TestVMBreakStopsLoop(t *testing.T) {
	src := `
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
	vm := mustCompile(t, src, nil)
	out, err := vm.Invoke("demo")
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if out.AsInt() != 3 {
		t.Fatalf("expected total 3, got %d", out.AsInt())
	}
}

func TestVMContinueSkipsIteration(t *testing.T) {
	src := `
func demo() {
  var nums = [1, 2, 3, 4]
  var total = 0
  for n in nums {
    if (n % 2 == 0) {
      continue
    }
    total = total + n
  }
  return total
}
`
	vm := mustCompile(t, src, nil)
	out, err := vm.Invoke("demo")
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if out.AsInt() != 4 {
		t.Fatalf("expected total 4, got %d", out.AsInt())
	}
}

func TestBreakOutsideLoopRuntimeError(t *testing.T) {
	src := `
func demo() {
  break
  return null
}
`
	vm := mustCompile(t, src, nil)
	if _, err := vm.Invoke("demo"); err == nil {
		t.Fatalf("expected error for break outside loop")
	} else if !strings.Contains(err.Error(), "break outside of loop") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnaryOperators(t *testing.T) {
	src := `
func demo() {
  var x = 2
  var flag = false
  x++
  return { neg: -x, pos: +x, not: !flag }
}
`
	vm := mustCompile(t, src, nil)
	out, err := vm.Invoke("demo")
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	obj := out.Obj
	if obj["neg"].AsInt() != -3 {
		t.Fatalf("expected neg=-3, got %d", obj["neg"].AsInt())
	}
	if obj["pos"].AsInt() != 3 {
		t.Fatalf("expected pos=3, got %d", obj["pos"].AsInt())
	}
	if !obj["not"].AsBool() {
		t.Fatalf("expected not=true")
	}
}

func TestObjectIteration(t *testing.T) {
	src := `
func demo() {
  var obj = { a: 1, b: 2, c: 3 }
  var total = 0
  for key in obj {
    total = total + obj[key]
  }
  return total
}
`
	vm := mustCompile(t, src, nil)
	out, err := vm.Invoke("demo")
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if out.AsInt() != 6 {
		t.Fatalf("expected sum 6, got %d", out.AsInt())
	}
}

func TestCompoundAssignments(t *testing.T) {
	src := `
func demo() {
  var x = 1
  x += 2
  x -= 1
  x++
  x--
  var arr = [1, 2]
  arr[0] += 3
  var obj = { a: 1 }
  obj.a += 4
  return { x: x, arr: arr[0], obj: obj.a }
}
`
	vm := mustCompile(t, src, nil)
	out, err := vm.Invoke("demo")
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	obj := out.Obj
	if obj["x"].AsInt() != 2 {
		t.Fatalf("expected x=2, got %d", obj["x"].AsInt())
	}
	if obj["arr"].AsInt() != 4 {
		t.Fatalf("expected arr=4, got %d", obj["arr"].AsInt())
	}
	if obj["obj"].AsInt() != 5 {
		t.Fatalf("expected obj.a=5, got %d", obj["obj"].AsInt())
	}
}
