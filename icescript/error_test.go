package icescript

import (
	"strings"
	"testing"
)

func TestRuntimeErrorStackTrace(t *testing.T) {
	src := `
func blow() {
  var z = 0
  return 10 / z
}

func main() {
  return blow()
}
`

	vm := NewVM()
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if _, err := vm.Invoke("main"); err == nil {
		t.Fatalf("expected runtime error")
	} else {
		msg := err.Error()
		if !strings.Contains(msg, "division by zero") {
			t.Fatalf("expected division by zero in error, got %q", msg)
		}
		if !strings.Contains(msg, "at blow") || !strings.Contains(msg, "at main") {
			t.Fatalf("expected stack trace in error, got %q", msg)
		}
	}
}

func TestCompileError(t *testing.T) {
	vm := NewVM()
	err := vm.Compile("func bad(\n")
	if err == nil {
		t.Fatalf("expected compile error")
	}
	if !strings.Contains(err.Error(), "expected ')'") {
		t.Fatalf("unexpected compile error: %v", err)
	}
}
