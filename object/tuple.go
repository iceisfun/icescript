package object

import (
	"bytes"
	"strings"
)

// Tuple represents a multi-value return from a function.
// It is primarily used internally by the VM and compiler for destructuring.
// If treated as a scalar, it behaves like its first element (Elements[0]).
type Tuple struct {
	Elements []Object
}

func (t *Tuple) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range t.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("(")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString(")")

	return out.String()
}

func (t *Tuple) Type() ObjectType { return TUPLE_OBJ }

// Delegated Methods for Scalar Compatibility (Case B)
// When a Tuple is assigned to a single variable or used in an expression,
// it should behave as its first element.

func (t *Tuple) first() Object {
	if len(t.Elements) > 0 {
		return t.Elements[0]
	}
	return NullObj
}

func (t *Tuple) AsFloat() (float64, bool) {
	return t.first().AsFloat()
}

func (t *Tuple) AsInt() (int64, bool) {
	return t.first().AsInt()
}

func (t *Tuple) AsString() (string, bool) {
	return t.first().AsString()
}

func (t *Tuple) AsBool() (bool, bool) {
	return t.first().AsBool()
}
