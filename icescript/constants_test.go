package icescript

import (
	"strings"
	"testing"
)

func TestConstantsProvideValues(t *testing.T) {
	vm := NewVM()
	if err := vm.SetConstants(map[string]any{"AncientTunnels": 65}); err != nil {
		t.Fatalf("SetConstants failed: %v", err)
	}
	const src = `
func demo() {
  return AncientTunnels
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	out, err := vm.Invoke("demo")
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if out.AsInt() != 65 {
		t.Fatalf("expected 65, got %d", out.AsInt())
	}
	if out.String() != "65::AncientTunnels" {
		t.Fatalf("expected annotated string, got %q", out.String())
	}
}

func TestConstantsImmutable(t *testing.T) {
	vm := NewVM()
	_ = vm.SetConstants(map[string]any{"AncientTunnels": 65})
	const src = `
func demo() {
  AncientTunnels = 1
  return null
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if _, err := vm.Invoke("demo"); err == nil {
		t.Fatalf("expected error when assigning constant")
	} else if want := "cannot assign to constant \"AncientTunnels\""; !strings.Contains(err.Error(), want) {
		t.Fatalf("expected %q in error, got %v", want, err)
	}
}

func TestConstantObjectImmutable(t *testing.T) {
	vm := NewVM()
	_ = vm.SetConstants(map[string]any{"Config": map[string]any{"answer": 42}})
	const src = `
func demo() {
  Config.answer = 1
  return null
}
`
	if err := vm.Compile(src); err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if _, err := vm.Invoke("demo"); err == nil {
		t.Fatalf("expected error when mutating constant object")
	} else if want := "cannot modify constant \"Config\""; !strings.Contains(err.Error(), want) {
		t.Fatalf("expected %q in error, got %v", want, err)
	}
}

func TestSetConstantsConversionError(t *testing.T) {
	vm := NewVM()
	if err := vm.SetConstants(map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatalf("expected conversion error for unsupported type")
	}
}
