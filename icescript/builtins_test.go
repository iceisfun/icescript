package icescript

import (
	"math"
	"testing"
)

func TestBuiltinsProduceExpectedValues(t *testing.T) {
	src := `
func demo() {
  var obj = {
    sqrt: sqrt(9),
    dist: distance(0, 0, 3, 4),
    sin0: sin(0),
    cos0: cos(0),
    atan1: atan(1),
    abs: abs(-5),
    lenStr: len("hello"),
    lenArr: len([1, 2, 3]),
    contains: contains("ell", "hello"),
    lower: lower("HELLO"),
    upper: upper("hello"),
    trim: trim("  spaced  "),
  }
  return obj
}
`
	vm := mustCompile(t, src, nil)
	out, err := vm.Invoke("demo")
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	obj := out.Obj
	if math.Abs(obj["sqrt"].AsFloat()-3) > 1e-9 {
		t.Fatalf("expected sqrt=3, got %v", obj["sqrt"].AsFloat())
	}
	if math.Abs(obj["dist"].AsFloat()-5) > 1e-9 {
		t.Fatalf("expected dist=5, got %v", obj["dist"].AsFloat())
	}
	if obj["abs"].AsInt() != 5 {
		t.Fatalf("expected abs=5, got %d", obj["abs"].AsInt())
	}
	if obj["lenStr"].AsInt() != 5 {
		t.Fatalf("expected lenStr=5, got %d", obj["lenStr"].AsInt())
	}
	if obj["lenArr"].AsInt() != 3 {
		t.Fatalf("expected lenArr=3, got %d", obj["lenArr"].AsInt())
	}
	if !obj["contains"].AsBool() {
		t.Fatalf("expected contains=true")
	}
	if obj["lower"].String() != "hello" {
		t.Fatalf("expected lower=hello, got %s", obj["lower"].String())
	}
	if obj["upper"].String() != "HELLO" {
		t.Fatalf("expected upper=HELLO, got %s", obj["upper"].String())
	}
	if obj["trim"].String() != "spaced" {
		t.Fatalf("expected trim=spaced, got %s", obj["trim"].String())
	}
}

func TestBuiltinSleep(t *testing.T) {
	vm := mustCompile(t, "func demo() { sleep(0); return null }", nil)
	if _, err := vm.Invoke("demo"); err != nil {
		t.Fatalf("sleep failed: %v", err)
	}
}
