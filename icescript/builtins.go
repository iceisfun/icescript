package icescript

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func installBuiltins(vm *VM) {
	builtins := map[string]HostFunc{
		"sqrt":     builtinSqrt,
		"distance": builtinDistance,
		"sin":      builtinSin,
		"cos":      builtinCos,
		"atan":     builtinAtan,
		"abs":      builtinAbs,
		"len":      builtinLen,
		"contains": builtinContains,
		"lower":    builtinLower,
		"upper":    builtinUpper,
		"trim":     builtinTrim,
		"sleep":    builtinSleep,
	}
	for name, fn := range builtins {
		vm.hostFuncs[name] = fn
	}
}

func builtinSqrt(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("sqrt", args, 1); err != nil {
		return VNull(), err
	}
	return VFloat(math.Sqrt(args[0].AsFloat())), nil
}

func builtinDistance(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("distance", args, 4); err != nil {
		return VNull(), err
	}
	x1 := args[0].AsFloat()
	y1 := args[1].AsFloat()
	x2 := args[2].AsFloat()
	y2 := args[3].AsFloat()
	return VFloat(math.Hypot(x2-x1, y2-y1)), nil
}

func builtinSin(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("sin", args, 1); err != nil {
		return VNull(), err
	}
	return VFloat(math.Sin(args[0].AsFloat())), nil
}

func builtinCos(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("cos", args, 1); err != nil {
		return VNull(), err
	}
	return VFloat(math.Cos(args[0].AsFloat())), nil
}

func builtinAtan(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("atan", args, 1); err != nil {
		return VNull(), err
	}
	return VFloat(math.Atan(args[0].AsFloat())), nil
}

func builtinAbs(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("abs", args, 1); err != nil {
		return VNull(), err
	}
	v := args[0]
	switch v.Kind {
	case FloatKind:
		return VFloat(math.Abs(v.AsFloat())), nil
	default:
		iv := v.AsInt()
		if iv < 0 {
			iv = -iv
		}
		return VInt(iv), nil
	}
}

func builtinLen(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("len", args, 1); err != nil {
		return VNull(), err
	}
	v := args[0]
	switch v.Kind {
	case StringKind:
		return VInt(int64(len(v.S))), nil
	case ArrayKind:
		return VInt(int64(len(v.Arr))), nil
	case ObjectKind:
		return VInt(int64(len(v.Obj))), nil
	default:
		return VInt(0), nil
	}
}

func builtinContains(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("contains", args, 2); err != nil {
		return VNull(), err
	}
	needle := args[0].String()
	haystack := args[1].String()
	return VBool(strings.Contains(haystack, needle)), nil
}

func builtinLower(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("lower", args, 1); err != nil {
		return VNull(), err
	}
	return VString(strings.ToLower(args[0].String())), nil
}

func builtinUpper(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("upper", args, 1); err != nil {
		return VNull(), err
	}
	return VString(strings.ToUpper(args[0].String())), nil
}

func builtinTrim(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("trim", args, 1); err != nil {
		return VNull(), err
	}
	return VString(strings.TrimSpace(args[0].String())), nil
}

func builtinSleep(_ *VM, args []Value) (Value, error) {
	if err := expectArgCount("sleep", args, 1); err != nil {
		return VNull(), err
	}
	duration := time.Duration(args[0].AsFloat() * float64(time.Millisecond))
	if duration < 0 {
		duration = 0
	}
	time.Sleep(duration)
	return VNull(), nil
}

func expectArgCount(name string, args []Value, want int) error {
	if len(args) != want {
		return fmt.Errorf("%s expects %d args, got %d", name, want, len(args))
	}
	return nil
}
