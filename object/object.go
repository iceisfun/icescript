package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"strconv"
	"strings"

	"math/rand"
	"time"

	"github.com/iceisfun/icescript/ast"
)

type BuiltinContext interface {
	Rand() *rand.Rand
	Now() time.Time
	Writer() io.Writer
	Get(k string) (any, bool)
	Set(k string, v any)
	PrintPrefix() string
}

type ObjectType string

const (
	INTEGER_OBJ           = "INTEGER"
	FLOAT_OBJ             = "FLOAT"
	BOOLEAN_OBJ           = "BOOLEAN"
	NULL_OBJ              = "NULL"
	RETURN_VALUE_OBJ      = "RETURN_VALUE"
	ERROR_OBJ             = "ERROR"
	FUNCTION_OBJ          = "FUNCTION"
	STRING_OBJ            = "STRING"
	BUILTIN_OBJ           = "BUILTIN"
	ARRAY_OBJ             = "ARRAY"
	HASH_OBJ              = "HASH"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION"
	CLOSURE_OBJ           = "CLOSURE"
)

type Object interface {
	Type() ObjectType
	Inspect() string

	// Type Conversion Helpers (Best Effort)
	AsFloat() (float64, bool)
	AsInt() (int64, bool)
	AsString() (string, bool)
	AsBool() (bool, bool)
}

type Hashable interface {
	HashKey() HashKey
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (i *Integer) AsFloat() (float64, bool) { return float64(i.Value), true }
func (i *Integer) AsInt() (int64, bool)     { return i.Value, true }
func (i *Integer) AsString() (string, bool) { return fmt.Sprintf("%d", i.Value), true }
func (i *Integer) AsBool() (bool, bool)     { return i.Value != 0, true }

type Float struct {
	Value float64
}

func (f *Float) Inspect() string  { return fmt.Sprintf("%f", f.Value) }
func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) HashKey() HashKey {
	return HashKey{Type: f.Type(), Value: uint64(f.Value)} // Note: Floats as map keys is risky
}

func (f *Float) AsFloat() (float64, bool) { return f.Value, true }
func (f *Float) AsInt() (int64, bool)     { return int64(f.Value), true }
func (f *Float) AsString() (string, bool) { return fmt.Sprintf("%f", f.Value), true }
func (f *Float) AsBool() (bool, bool)     { return f.Value != 0.0, true }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}

func (b *Boolean) AsFloat() (float64, bool) {
	if b.Value {
		return 1.0, true
	}
	return 0.0, true
}
func (b *Boolean) AsInt() (int64, bool) {
	if b.Value {
		return 1, true
	}
	return 0, true
}
func (b *Boolean) AsString() (string, bool) { return fmt.Sprintf("%t", b.Value), true }
func (b *Boolean) AsBool() (bool, bool)     { return b.Value, true }

type Null struct{}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL_OBJ }

func (n *Null) AsFloat() (float64, bool) { return 0, false }
func (n *Null) AsInt() (int64, bool)     { return 0, false }
func (n *Null) AsString() (string, bool) { return "", false }
func (n *Null) AsBool() (bool, bool)     { return false, false }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

func (rv *ReturnValue) AsFloat() (float64, bool) { return 0, false }
func (rv *ReturnValue) AsInt() (int64, bool)     { return 0, false }
func (rv *ReturnValue) AsString() (string, bool) { return "", false }
func (rv *ReturnValue) AsBool() (bool, bool)     { return false, false }

type Error struct {
	Message string
}

func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
func (e *Error) Type() ObjectType { return ERROR_OBJ }

func (e *Error) AsFloat() (float64, bool) { return 0, false }
func (e *Error) AsInt() (int64, bool)     { return 0, false }
func (e *Error) AsString() (string, bool) { return "", false }
func (e *Error) AsBool() (bool, bool)     { return false, false }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        interface{} // Placeholder for environment if needed for Eval walker, but we are doing Bytecode
}

func (f *Function) Inspect() string  { return "fn(...)" }
func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

func (f *Function) AsFloat() (float64, bool) { return 0, false }
func (f *Function) AsInt() (int64, bool)     { return 0, false }
func (f *Function) AsString() (string, bool) { return "", false }
func (f *Function) AsBool() (bool, bool)     { return false, false }

type String struct {
	Value string
}

func (s *String) Inspect() string  { return s.Value }
func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

func (s *String) AsFloat() (float64, bool) {
	v, err := strconv.ParseFloat(s.Value, 64)
	return v, err == nil
}
func (s *String) AsInt() (int64, bool) {
	v, err := strconv.ParseInt(s.Value, 10, 64)
	return v, err == nil
}
func (s *String) AsString() (string, bool) { return s.Value, true }
func (s *String) AsBool() (bool, bool)     { return s.Value == "true", true }

type BuiltinFunction func(ctx BuiltinContext, args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Inspect() string  { return "builtin function" }
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }

func (b *Builtin) AsFloat() (float64, bool) { return 0, false }
func (b *Builtin) AsInt() (int64, bool)     { return 0, false }
func (b *Builtin) AsString() (string, bool) { return "", false }
func (b *Builtin) AsBool() (bool, bool)     { return false, false }

type Array struct {
	Elements []Object
}

func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}
func (ao *Array) Type() ObjectType { return ARRAY_OBJ }

func (ao *Array) AsFloat() (float64, bool) { return 0, false }
func (ao *Array) AsInt() (int64, bool)     { return 0, false }
func (ao *Array) AsString() (string, bool) { return "", false }
func (ao *Array) AsBool() (bool, bool)     { return false, false }

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
func (h *Hash) Type() ObjectType { return HASH_OBJ }

func (h *Hash) AsFloat() (float64, bool) { return 0, false }
func (h *Hash) AsInt() (int64, bool)     { return 0, false }
func (h *Hash) AsString() (string, bool) { return "", false }
func (h *Hash) AsBool() (bool, bool)     { return false, false }

type CompiledFunction struct {
	Instructions  []byte
	NumLocals     int
	NumParameters int
	SourceMap     map[int]int
	Name          string
}

func (cf *CompiledFunction) Inspect() string  { return fmt.Sprintf("CompiledFunction[%p]", cf) }
func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }

func (cf *CompiledFunction) AsFloat() (float64, bool) { return 0, false }
func (cf *CompiledFunction) AsInt() (int64, bool)     { return 0, false }
func (cf *CompiledFunction) AsString() (string, bool) { return "", false }
func (cf *CompiledFunction) AsBool() (bool, bool)     { return false, false }

type Panic struct {
	Message string
}

func (p *Panic) Inspect() string  { return "PANIC: " + p.Message }
func (p *Panic) Type() ObjectType { return ERROR_OBJ }

func (p *Panic) AsFloat() (float64, bool) { return 0, false }
func (p *Panic) AsInt() (int64, bool)     { return 0, false }
func (p *Panic) AsString() (string, bool) { return "", false }
func (p *Panic) AsBool() (bool, bool)     { return false, false }

type Closure struct {
	Fn   *CompiledFunction
	Free []Object
}

func (c *Closure) Inspect() string  { return fmt.Sprintf("Closure[%p]", c) }
func (c *Closure) Type() ObjectType { return CLOSURE_OBJ }

func (c *Closure) AsFloat() (float64, bool) { return 0, false }
func (c *Closure) AsInt() (int64, bool)     { return 0, false }
func (c *Closure) AsString() (string, bool) { return "", false }
func (c *Closure) AsBool() (bool, bool)     { return false, false }

// Helpers
func NativeBoolToBooleanObject(input bool) *Boolean {
	if input {
		return True
	}
	return False
}
