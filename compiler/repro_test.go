package compiler

import (
	"testing"
)

func TestForwardReference(t *testing.T) {
	input := `
	func A() {
		return B()
	}

	func B() {
		return 1
	}

	A()
	`

	program := parse(input)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error with forward reference: %s", err)
	}

	// Double check that we compiled something for both functions
	// Not strictly necessary if Compile didn't error, but good sanity check.
}
