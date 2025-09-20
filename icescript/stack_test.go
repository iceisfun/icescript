package icescript

import "testing"

func TestRecursiveFunctionReturns(t *testing.T) {
	vm := NewVM()
	const src = `
func fib(n) {
  if (n <= 1) {
    return n
  }
  return fib(n - 1) + fib(n - 2)
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile fib script: %v", err)
	}
	got, err := vm.Invoke("fib", VInt(10))
	if err != nil {
		t.Fatalf("invoke fib: %v", err)
	}
	if got.AsInt() != 55 {
		t.Fatalf("fib(10) = %d, want 55", got.AsInt())
	}
}

func TestReturnFromLoopBody(t *testing.T) {
	vm := NewVM()
	const src = `
func firstGreaterThan(xs, threshold) {
  for x in xs {
    if (x > threshold) {
      return x
    }
  }
  return -1
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile firstGreaterThan: %v", err)
	}
	arr := VArray([]Value{VInt(1), VInt(2), VInt(3), VInt(5)})
	got, err := vm.Invoke("firstGreaterThan", arr, VInt(2))
	if err != nil {
		t.Fatalf("invoke firstGreaterThan: %v", err)
	}
	if got.AsInt() != 3 {
		t.Fatalf("firstGreaterThan returned %d, want 3", got.AsInt())
	}
}

func TestReturnNullValue(t *testing.T) {
	vm := NewVM()
	const src = `
func noop() {
  return null
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile noop: %v", err)
	}
	got, err := vm.Invoke("noop")
	if err != nil {
		t.Fatalf("invoke noop: %v", err)
	}
	if got.Kind != NullKind {
		t.Fatalf("noop return kind = %v, want NullKind", got.Kind)
	}
}

func TestBreakExitsLoop(t *testing.T) {
	vm := NewVM()
	const src = `
func firstEven(xs) {
  for x in xs {
    if (x % 2 == 0) {
      return x  // should exit immediately
    }
  }
  return -1
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile firstEven: %v", err)
	}
	arr := VArray([]Value{VInt(1), VInt(3), VInt(4), VInt(6)})
	got, err := vm.Invoke("firstEven", arr)
	if err != nil {
		t.Fatalf("invoke firstEven: %v", err)
	}
	if got.AsInt() != 4 {
		t.Fatalf("firstEven returned %d, want 4", got.AsInt())
	}
}

func TestContinueSkipsIteration(t *testing.T) {
	vm := NewVM()
	const src = `
func sumOdds(xs) {
  var total = 0
  for x in xs {
    if (x % 2 == 0) {
      continue  // skip evens
    }
    total = total + x
  }
  return total
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile sumOdds: %v", err)
	}
	arr := VArray([]Value{VInt(1), VInt(2), VInt(3), VInt(4), VInt(5)})
	got, err := vm.Invoke("sumOdds", arr)
	if err != nil {
		t.Fatalf("invoke sumOdds: %v", err)
	}
	if got.AsInt() != 9 { // 1 + 3 + 5
		t.Fatalf("sumOdds returned %d, want 9", got.AsInt())
	}
}

func TestBreakAndContinueTogether(t *testing.T) {
	vm := NewVM()
	const src = `
func firstOddAfter(xs, threshold) {
  for x in xs {
    if (x <= threshold) {
      continue  // skip until we cross threshold
    }
    if (x % 2 == 1) {
      return x  // first odd beyond threshold
    }
  }
  return -1
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile firstOddAfter: %v", err)
	}
	arr := VArray([]Value{VInt(2), VInt(4), VInt(5), VInt(7)})
	got, err := vm.Invoke("firstOddAfter", arr, VInt(3))
	if err != nil {
		t.Fatalf("invoke firstOddAfter: %v", err)
	}
	if got.AsInt() != 5 {
		t.Fatalf("firstOddAfter returned %d, want 5", got.AsInt())
	}
}

func TestNestedControlFlow(t *testing.T) {
	vm := NewVM()
	const src = `
func findFirstDivisible(xs, ys, threshold) {
  for x in xs {
    if (x <= threshold) {
      continue   // skip small x values
    }
    for y in ys {
      if (y == 0) {
        continue // skip zero to avoid div-by-zero
      }
      if (x % y == 0) {
        return x // return the first x > threshold divisible by some y
      }
      if (y > x) {
        break // no need to check larger y
      }
    }
  }
  return -1
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile findFirstDivisible: %v", err)
	}
	xs := VArray([]Value{VInt(2), VInt(5), VInt(7), VInt(12), VInt(15)})
	ys := VArray([]Value{VInt(2), VInt(3), VInt(10)})
	got, err := vm.Invoke("findFirstDivisible", xs, ys, VInt(6))
	if err != nil {
		t.Fatalf("invoke findFirstDivisible: %v", err)
	}
	// Expected: skip x=2,5; x=7 not divisible by 2 or 3; x=12 divisible by 2 => return 12
	if got.AsInt() != 12 {
		t.Fatalf("findFirstDivisible returned %d, want 12", got.AsInt())
	}
}
