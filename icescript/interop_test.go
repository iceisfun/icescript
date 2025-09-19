package icescript

import (
	"reflect"
	"testing"
)

type interopPlayer struct {
	Name   string
	Life   int
	Status string `script:"state"`
	Scores []int
	Tags   map[string]string
}

func TestValueRoundTripStruct(t *testing.T) {
	original := interopPlayer{
		Name:   "Hero",
		Life:   42,
		Status: "active",
		Scores: []int{1, 2, 3},
		Tags: map[string]string{
			"class": "mage",
		},
	}

	val := MustVFromGo(original)

	var out interopPlayer
	if err := val.ToGo(&out); err != nil {
		t.Fatalf("ToGo failed: %v", err)
	}

	if !reflect.DeepEqual(original, out) {
		t.Fatalf("round trip mismatch:\noriginal: %#v\nresult: %#v", original, out)
	}
}

func TestValueRoundTripMapSlice(t *testing.T) {
	src := map[string]any{
		"name":    "crate",
		"weights": []float64{1.5, 2.5},
		"flags":   []bool{true, false},
	}

	val := MustVFromGo(src)

	var out map[string]any
	if err := val.ToGo(&out); err != nil {
		t.Fatalf("ToGo failed: %v", err)
	}

	if got := out["name"]; got != "crate" {
		t.Fatalf("expected name 'crate', got %v", got)
	}

	weights, ok := out["weights"].([]any)
	if !ok || len(weights) != 2 {
		t.Fatalf("expected weights slice of len 2, got %#v", out["weights"])
	}
}

func TestToGoRequiresPointer(t *testing.T) {
	val := VInt(1)
	if err := val.ToGo(nil); err == nil {
		t.Fatalf("expected error when passing nil")
	}
	if err := val.ToGo(1); err == nil {
		t.Fatalf("expected error when passing non-pointer")
	}
}
