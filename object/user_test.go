package object

import (
	"fmt"
	"testing"
)

type stringer struct{}

func (s stringer) String() string { return "I am stringer" }

type plain struct{}

func TestUserObject(t *testing.T) {
	// 1. Test Stringer
	u1 := &User{Value: stringer{}}
	if u1.Type() != USER_OBJ {
		t.Errorf("expected USER_OBJ, got %s", u1.Type())
	}
	if u1.Inspect() != "I am stringer" {
		t.Errorf("expected 'I am stringer', got %q", u1.Inspect())
	}

	// 2. Test Plain
	u2 := &User{Value: plain{}}
	expected := " <user object.plain>" // Allow slight variation or exact match
	// The implementation uses fmt.Sprintf("<user %T>", u.Value)
	expected = fmt.Sprintf("<user %T>", plain{})
	if u2.Inspect() != expected {
		t.Errorf("expected %q, got %q", expected, u2.Inspect())
	}

	// 3. Test Conversions (should fail/default)
	if _, ok := u1.AsInt(); ok {
		t.Errorf("AsInt should be false")
	}
	if _, ok := u1.AsFloat(); ok {
		t.Errorf("AsFloat should be false")
	}
	if _, ok := u1.AsBool(); ok {
		t.Errorf("AsBool should be false")
	}
	if str, ok := u1.AsString(); !ok || str != "I am stringer" {
		t.Errorf("AsString should return inspect value")
	}
}
