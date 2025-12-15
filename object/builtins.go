package object

import "fmt"

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}

			switch arg := args[0].(type) {
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			default:
				return &Error{Message: fmt.Sprintf("argument to `len` not supported, got %s", args[0].Type())}
			}
		}},
	},
	{
		"print",
		&Builtin{Fn: func(args ...Object) Object {
			out := []interface{}{}
			for _, arg := range args {
				out = append(out, arg.Inspect())
			}
			fmt.Println(out...)
			return NullObj
		}},
	},
	{
		"panic",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}
			return &Panic{Message: args[0].Inspect()}
		}},
	},
	{
		"push",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return &Error{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=2", len(args))}
			}

			if args[0].Type() != ARRAY_OBJ {
				return &Error{Message: fmt.Sprintf("argument to `push` must be ARRAY, got %s", args[0].Type())}
			}

			arr := args[0].(*Array)
			arr.Elements = append(arr.Elements, args[1])
			// Return successful array or null? Standard is usually returning the new length or the array.
			// Let's return the array for chaining or just consistency.
			return arr
		}},
	},
	{
		"keys",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args))}
			}

			hash, ok := args[0].(*Hash)
			if !ok {
				return &Error{Message: fmt.Sprintf("argument to `keys` must be HASH, got %s", args[0].Type())}
			}

			elements := []Object{}
			for _, pair := range hash.Pairs {
				elements = append(elements, pair.Key)
			}
			return &Array{Elements: elements}
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
