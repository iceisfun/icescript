package object

import "fmt"

type User struct {
	Value interface{}
}

func (u *User) Inspect() string {
	if stringer, ok := u.Value.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("<user %T>", u.Value)
}

func (u *User) Type() ObjectType { return USER_OBJ }

func (u *User) Equal(other Object) (bool, error) {
	if other.Type() != USER_OBJ {
		return false, fmt.Errorf("type mismatch: %s == %s", u.Type(), other.Type())
	}
	// otherUser := other.(*User) // Unused, we pass 'other' to delegated Equal

	if eqObj, ok := u.Value.(ObjectEqual); ok {
		// We need to pass the *inner* value to the inner Equal method if it expects that.
		// However, ObjectEqual.Equal expects an storage.Object.
		// The inner implementation (like EqualUser in tests) likely expects another inner implementation.
		// The interface signature is Equal(other Object) (bool, error).
		// So we pass 'other' (the *User wrapper) or 'other.Value'?
		// If EqualUser implements ObjectEqual, it's method is Equal(other Object).
		// If we pass the wrapper *User, EqualUser must unwrap it.
		// Let's check the SOW or intent. SOW: "If both operands ... Implement ObjectEqual → invoke Equal".
		// The operands are *User.
		// So *User should implement ObjectEqual.
		// And it delegates.
		// But EqualUser (inner) might not know about *User wrapper if it's a host object.
		// Actually, if the Host implements ObjectEqual on the struct, it expects to compare against another struct of same type.
		// It shouldn't need to know about icescript.Object wrappers if possible.
		// But the specific signature `Equal(other Object)` forces dependency on `Object`.
		// So the Host object MUST import `object` package.
		// In that case, it knows about `Object`.
		// My test implementation `EqualUser` unwraps `*object.User`.
		// So passing `other` (*User) is correct IF the implementation handles it.
		// Let's assume the implementation handles *User unwrapping as done in the test.
		return eqObj.Equal(other)
	}
	return false, fmt.Errorf("equality not supported for type: %s", u.Type())
}

// User objects are NOT hashable by default to prevent complex key behavior
// func (u *User) HashKey() HashKey { ... }

func (u *User) AsFloat() (float64, bool) { return 0, false }
func (u *User) AsInt() (int64, bool)     { return 0, false }
func (u *User) AsString() (string, bool) { return u.Inspect(), true }
func (u *User) AsBool() (bool, bool)     { return false, false }
