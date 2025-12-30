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

// User objects are NOT hashable by default to prevent complex key behavior
// func (u *User) HashKey() HashKey { ... }

func (u *User) AsFloat() (float64, bool) { return 0, false }
func (u *User) AsInt() (int64, bool)     { return 0, false }
func (u *User) AsString() (string, bool) { return u.Inspect(), true }
func (u *User) AsBool() (bool, bool)     { return false, false }
