package icescript

import "testing"

// TestShadowing ensures that variables declared inside a loop
// shadow outer variables instead of leaking/overwriting them.
func TestShadowing(t *testing.T) {
	vm := NewVM()
	const src = `
func shadow(xs) {
  var x = 1
  for y in xs {
    var x = y  // should shadow only inside loop body
  }
  return x    // should still be 1
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile shadow script: %v", err)
	}
	arr := VArray([]Value{VInt(5), VInt(6), VInt(7)})
	got, err := vm.Invoke("shadow", arr)
	if err != nil {
		t.Fatalf("invoke shadow: %v", err)
	}
	if got.AsInt() != 1 {
		t.Fatalf("shadow returned %d, want 1 (outer x should not be overwritten)", got.AsInt())
	}
}
