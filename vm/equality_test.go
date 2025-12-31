package vm

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/iceisfun/icescript/compiler"
	"github.com/iceisfun/icescript/lexer"
	"github.com/iceisfun/icescript/object"
	"github.com/iceisfun/icescript/parser"
)

// Mock Equality Object
type EqualUser struct {
	ID int
}

func (e *EqualUser) Type() object.ObjectType  { return object.USER_OBJ }
func (e *EqualUser) Inspect() string          { return fmt.Sprintf("User(%d)", e.ID) }
func (e *EqualUser) AsFloat() (float64, bool) { return 0, false }
func (e *EqualUser) AsInt() (int64, bool)     { return 0, false }
func (e *EqualUser) AsString() (string, bool) { return "", false }
func (e *EqualUser) AsBool() (bool, bool)     { return false, false }

// Implement ObjectEqual
func (e *EqualUser) Equal(other object.Object) (bool, error) {
	o, ok := other.(*object.User)
	if !ok {
		return false, fmt.Errorf("type mismatch in Equal")
	}
	val, ok := o.Value.(*EqualUser)
	if !ok {
		return false, fmt.Errorf("value mismatch in Equal")
	}
	return e.ID == val.ID, nil
}

// Mock No-Equality Object
type NoEqualUser struct {
	ID int
}

func (n *NoEqualUser) Type() object.ObjectType  { return object.USER_OBJ }
func (n *NoEqualUser) Inspect() string          { return fmt.Sprintf("NoEqual(%d)", n.ID) }
func (n *NoEqualUser) AsFloat() (float64, bool) { return 0, false }
func (n *NoEqualUser) AsInt() (int64, bool)     { return 0, false }
func (n *NoEqualUser) AsString() (string, bool) { return "", false }
func (n *NoEqualUser) AsBool() (bool, bool)     { return false, false }

func TestEquality(t *testing.T) {
	tests := []struct {
		desc     string
		input    string
		setup    func(*VM)
		expected interface{} // bool for result, or string for error
	}{
		// Primitive Equality (Preserved)
		{"Int Equal", "1 == 1", nil, true},
		{"Int Not Equal", "1 != 2", nil, true},
		{"String Equal", `"foo" == "foo"`, nil, true},
		{"String Not Equal", `"foo" != "bar"`, nil, true},
		{"Bool Equal", "true == true", nil, true},
		{"Null Equal", "null == null", nil, true},
		{"Cross-primitive false", "1 == \"1\"", nil, false}, // Existing behavior preserved

		// Internal Types Error
		{"Array Equality Error", "[1] == [1]", nil, "equality not supported for type: ARRAY"},
		{"Hash Equality Error", "{1:1} == {1:1}", nil, "equality not supported for type: HASH"},
		{"Function Equality Error", "func(){} == func(){}", nil, "equality not supported for type: CLOSURE"},

		// User Object With Equality
		{
			"User Explicit Equal",
			"u1 == u2",
			func(vm *VM) {
				vm.SetGlobal(0, &object.User{Value: &EqualUser{ID: 123}})
				vm.SetGlobal(1, &object.User{Value: &EqualUser{ID: 123}})
			},
			true,
		},
		{
			"User Explicit Not Equal",
			"u1 != u2",
			func(vm *VM) {
				vm.SetGlobal(0, &object.User{Value: &EqualUser{ID: 123}})
				vm.SetGlobal(1, &object.User{Value: &EqualUser{ID: 456}})
			},
			true,
		},

		// User Object Without Equality
		{
			"User No Support",
			"u1 == u2", // Changed from n1 == n2 to re-use u1/u2 indices 0,1
			func(vm *VM) {
				vm.SetGlobal(0, &object.User{Value: &NoEqualUser{ID: 1}})
				vm.SetGlobal(1, &object.User{Value: &NoEqualUser{ID: 1}})
			},
			"equality not supported for type: USER_OBJ",
		},

		// Cross Type Non-Primitive Errors
		{
			"User vs Int",
			"u1 == 1",
			func(vm *VM) {
				vm.SetGlobal(0, &object.User{Value: &EqualUser{ID: 1}})
			},
			"type mismatch: USER_OBJ == INTEGER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program := p.ParseProgram()

			c := compiler.New()
			// Only define used symbols to avoid index confusion, or just define consistently.
			// Reusing u1/u2 for all user object tests is simplest given we hardcode SetGlobal(0/1).
			c.SymbolTable().Define("u1")
			c.SymbolTable().Define("u2")

			err := c.Compile(program)
			if err != nil {
				t.Fatalf("compile error: %s", err)
			}

			vm := New(c.Bytecode())
			if tt.setup != nil {
				tt.setup(vm)
			}

			err = vm.Run(context.Background())

			// Check expected error
			if expectedErr, ok := tt.expected.(string); ok {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", expectedErr)
				} else if !strings.Contains(err.Error(), expectedErr) {
					t.Errorf("expected error containing %q, got %q", expectedErr, err.Error())
				}
				return
			}

			// Check success
			if err != nil {
				t.Fatalf("vm error: %s", err)
			}

			last := vm.LastPoppedStackElem()
			result, ok := last.(*object.Boolean)
			if !ok {
				t.Fatalf("expected boolean result, got %T", last)
			}
			if result.Value != tt.expected.(bool) {
				t.Errorf("expected %v, got %v", tt.expected, result.Value)
			}
		})
	}
}
