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

func (t *Tuple) Equal(other Object) (bool, error) {
	if other.Type() != TUPLE_OBJ {
		return false, nil // Not equal if types differ
	}

	o := other.(*Tuple)

	if len(t.Elements) != len(o.Elements) {
		return false, nil
	}

	for i, el := range t.Elements {
		// Use VM logic for comparison? Handled by Object logic ideally,
		// but Object interface doesn't expose generic Equals.
		// Comparison logic is usually in VM.
		// But here `Equal` is called by VM for complex types.

		// Problem: primitive types like Integer don't implement Equal usually.
		// VM handles primitives. ObjectEqual is for complex user types or extensive types.
		// If Tuple elements are primitives, we need to compare them.

		// If element implements ObjectEqual, use it because it will handle cross-comparisons if designed well ???
		// No, ObjectEqual is usually for UserObjs.
		// But we need to compare integers too.

		// Let's rely on string inspection as a fallback or handle primitives manually?
		// Better: Check types and values.

		if el.Type() != o.Elements[i].Type() {
			return false, nil
		}

		// Primitive checks
		switch el.Type() {
		case INTEGER_OBJ:
			if el.(*Integer).Value != o.Elements[i].(*Integer).Value {
				return false, nil
			}
		case FLOAT_OBJ:
			if el.(*Float).Value != o.Elements[i].(*Float).Value {
				return false, nil
			}
		case BOOLEAN_OBJ:
			if el.(*Boolean).Value != o.Elements[i].(*Boolean).Value {
				return false, nil
			}
		case STRING_OBJ:
			if el.(*String).Value != o.Elements[i].(*String).Value {
				return false, nil
			}
		case NULL_OBJ:
			continue
		case TUPLE_OBJ:
			eq, err := el.(*Tuple).Equal(o.Elements[i])
			if err != nil {
				return false, err
			}
			if !eq {
				return false, nil
			}
		default:
			// Fallback or Not Supported?
			// Use Inspect for deep equality?
			if el.Inspect() != o.Elements[i].Inspect() {
				return false, nil
			}
		}
	}

	return true, nil
}
