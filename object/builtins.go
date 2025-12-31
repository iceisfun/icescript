package object

import (
	"fmt"
	"math"
	"strings"
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 1 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}

			switch arg := args[0].(type) {
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			default:
				return &Critical{Message: fmt.Sprintf("argument to `len` not supported, got %s", args[0].Type())}
			}
		}},
	},
	{
		"print",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			out := []interface{}{}
			if prefix := ctx.PrintPrefix(); prefix != "" {
				out = append(out, prefix)
			}
			for _, arg := range args {
				out = append(out, arg.Inspect())
			}
			w := ctx.Writer()
			if w == nil {
				fmt.Println(out...)
			} else {
				fmt.Fprintln(w, out...)
			}
			return NullObj
		}},
	},
	{
		"panic",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 1 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}
			return &Panic{Message: args[0].Inspect()}
		}},
	},
	{
		"push",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 2 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=2", len(args))}
			}

			if args[0].Type() != ARRAY_OBJ {
				return &Critical{Message: fmt.Sprintf("argument to `push` must be ARRAY, got %s", args[0].Type())}
			}

			arr := args[0].(*Array)
			arr.Elements = append(arr.Elements, args[1])
			return arr
		}},
	},
	{
		"keys",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 1 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}

			hash, ok := args[0].(*Hash)
			if !ok {
				return &Critical{Message: fmt.Sprintf("argument to `keys` must be HASH, got %s", args[0].Type())}
			}

			elements := []Object{}
			for _, pair := range hash.Pairs {
				elements = append(elements, pair.Key)
			}
			return &Array{Elements: elements}
		}},
	},
	{
		"contains",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 2 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=2", len(args))}
			}

			switch container := args[0].(type) {
			case *Array:
				for _, el := range container.Elements {
					if el.Inspect() == args[1].Inspect() { // Simple equality check for now
						return True
					}
				}
				return False
			case *String:
				sub, ok := args[1].(*String)
				if !ok {
					return &Critical{Message: fmt.Sprintf("second argument to `contains` for STRING must be STRING, got %s", args[1].Type())}
				}
				if strings.Contains(container.Value, sub.Value) {
					return True
				}
				return False
			case *Hash:
				key, ok := args[1].(Hashable)
				if !ok {
					return &Critical{Message: fmt.Sprintf("unusable as hash key: %s", args[1].Type())}
				}
				if _, ok := container.Pairs[key.HashKey()]; ok {
					return True
				}
				return False
			default:
				return &Critical{Message: fmt.Sprintf("argument to `contains` not supported, got %s", args[0].Type())}
			}
		}},
	},
	{
		"distance",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 4 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=4", len(args))}
			}
			vals := make([]float64, 4)
			for i, arg := range args {
				switch v := arg.(type) {
				case *Integer:
					vals[i] = float64(v.Value)
				case *Float:
					vals[i] = v.Value
				default:
					return &Critical{Message: fmt.Sprintf("argument %d to `distance` must be number, got %s", i, arg.Type())}
				}
			}
			return &Float{Value: math.Hypot(vals[2]-vals[0], vals[3]-vals[1])}
		}},
	},
	{
		"hypot",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 4 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=4", len(args))}
			}
			vals := make([]float64, 4)
			for i, arg := range args {
				switch v := arg.(type) {
				case *Integer:
					vals[i] = float64(v.Value)
				case *Float:
					vals[i] = v.Value
				default:
					return &Critical{Message: fmt.Sprintf("argument %d to `hypot` must be number, got %s", i, arg.Type())}
				}
			}
			return &Float{Value: math.Hypot(vals[2]-vals[0], vals[3]-vals[1])}
		}},
	},
	{
		"sqrt",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 1 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}
			var val float64
			switch v := args[0].(type) {
			case *Integer:
				val = float64(v.Value)
			case *Float:
				val = v.Value
			default:
				return &Critical{Message: fmt.Sprintf("argument to `sqrt` must be number, got %s", args[0].Type())}
			}
			return &Float{Value: math.Sqrt(val)}
		}},
	},
	{
		"atan2",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 2 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=2", len(args))}
			}
			vals := make([]float64, 2)
			for i, arg := range args {
				switch v := arg.(type) {
				case *Integer:
					vals[i] = float64(v.Value)
				case *Float:
					vals[i] = v.Value
				default:
					return &Critical{Message: fmt.Sprintf("argument %d to `atan2` must be number, got %s", i, arg.Type())}
				}
			}
			return &Float{Value: math.Atan2(vals[0], vals[1])}
		}},
	},
	{
		"equalFold",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 2 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=2", len(args))}
			}
			s1, ok1 := args[0].(*String)
			s2, ok2 := args[1].(*String)
			if !ok1 || !ok2 {
				return &Critical{Message: "arguments to `equalFold` must be STRINGs"}
			}
			return NativeBoolToBooleanObject(strings.EqualFold(s1.Value, s2.Value))
		}},
	},
	{
		"seed",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 1 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}
			if ctx == nil || ctx.Rand() == nil {
				return &Critical{Message: "RNG not available in this context"}
			}

			var val int64
			switch v := args[0].(type) {
			case *Integer:
				val = v.Value
			default:
				return &Critical{Message: fmt.Sprintf("argument to `seed` must be INTEGER, got %s", args[0].Type())}
			}
			ctx.Rand().Seed(val)
			return NullObj
		}},
	},
	{
		"random",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 0 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=0", len(args))}
			}
			if ctx == nil || ctx.Rand() == nil {
				return &Critical{Message: "RNG not available in this context"}
			}
			return &Float{Value: ctx.Rand().Float64()}
		}},
	},
	{
		"randomInt",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 2 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=2", len(args))}
			}
			if ctx == nil || ctx.Rand() == nil {
				return &Critical{Message: "RNG not available in this context"}
			}

			var min int64
			switch v := args[0].(type) {
			case *Integer:
				min = v.Value
			default:
				return &Critical{Message: fmt.Sprintf("argument to `randomInt` must be INTEGER, got %s", args[0].Type())}
			}

			var max int64
			switch v := args[1].(type) {
			case *Integer:
				max = v.Value
			default:
				return &Critical{Message: fmt.Sprintf("argument to `randomInt` must be INTEGER, got %s", args[1].Type())}
			}

			if min >= max {
				return &Critical{Message: fmt.Sprintf("min (%d) must be less than max (%d)", min, max)}
			}

			diff := max - min
			return &Integer{Value: ctx.Rand().Int63n(diff) + min}
		}},
	},
	{
		"randomItem",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 1 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}
			if ctx == nil || ctx.Rand() == nil {
				return &Critical{Message: "RNG not available in this context"}
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return &Critical{Message: fmt.Sprintf("argument to `randomItem` must be ARRAY, got %s", args[0].Type())}
			}

			if len(arr.Elements) == 0 {
				return NullObj
			}

			idx := ctx.Rand().Intn(len(arr.Elements))
			return arr.Elements[idx]
		}},
	},
	{
		"now",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 0 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=0", len(args))}
			}
			if ctx == nil {
				return &Critical{Message: "Context not available"}
			}
			return &Integer{Value: ctx.Now().UnixMilli()}
		}},
	},
	{
		"since",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 1 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}
			if ctx == nil {
				return &Critical{Message: "Context not available"}
			}

			start, ok := args[0].(*Integer)
			if !ok {
				return &Critical{Message: fmt.Sprintf("argument to `since` must be INTEGER (timestamp), got %s", args[0].Type())}
			}

			now := ctx.Now().UnixMilli()
			return &Integer{Value: now - start.Value}
		}},
	},
	{
		"testMultiReturn",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 2 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=2", len(args))}
			}
			return &Tuple{Elements: []Object{args[0], args[1]}}
		}},
	},
	{
		"typeof",
		&Builtin{Fn: func(ctx BuiltinContext, args ...Object) Object {
			if len(args) != 1 {
				return &Critical{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}
			t := args[0].Type()
			var s string
			switch t {
			case INTEGER_OBJ:
				s = "integer"
			case FLOAT_OBJ:
				s = "float"
			case BOOLEAN_OBJ:
				s = "boolean"
			case NULL_OBJ:
				s = "null"
			case ERROR_OBJ:
				s = "error"
			case STRING_OBJ:
				s = "string"
			case BUILTIN_OBJ:
				s = "builtin"
			case ARRAY_OBJ:
				s = "array"
			case USER_OBJ:
				s = "user"
			case TUPLE_OBJ:
				s = "tuple"
			default:
				s = string(t)
			}
			return &String{Value: s}
		}},
	},
}

var NullObj = &Null{}
var True = &Boolean{Value: true}
var False = &Boolean{Value: false}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}
