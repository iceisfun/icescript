package icescript

import (
	"fmt"
	"math"
	"reflect"
)

var valueType = reflect.TypeOf(Value{})

// VFromGo converts a plain Go value into an icescript Value. Only simple POD
// types (numbers, strings, bools), slices/arrays, maps with string keys, and
// structs composed of those types are supported. Struct fields can use the
// `script` tag to rename or skip fields (use "-" to skip).
func VFromGo(v any) (Value, error) {
	if v == nil {
		return VNull(), nil
	}
	if val, ok := v.(Value); ok {
		return val, nil
	}
	return convertReflectValue(reflect.ValueOf(v))
}

// MustVFromGo converts a Go value into an icescript Value and panics on error.
func MustVFromGo(v any) Value {
	val, err := VFromGo(v)
	if err != nil {
		panic(err)
	}
	return val
}

func convertReflectValue(rv reflect.Value) (Value, error) {
	if !rv.IsValid() {
		return VNull(), nil
	}

	switch rv.Kind() {
	case reflect.Interface, reflect.Pointer:
		if rv.IsNil() {
			return VNull(), nil
		}
		return convertReflectValue(rv.Elem())

	case reflect.Bool:
		return VBool(rv.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return VInt(rv.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u := rv.Uint()
		if u > math.MaxInt64 {
			return Value{}, fmt.Errorf("cannot convert %v: value %d overflows int64", rv.Type(), u)
		}
		return VInt(int64(u)), nil

	case reflect.Float32, reflect.Float64:
		return VFloat(rv.Convert(reflect.TypeOf(float64(0))).Float()), nil

	case reflect.String:
		return VString(rv.String()), nil

	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			buf := make([]byte, rv.Len())
			reflect.Copy(reflect.ValueOf(buf), rv)
			return VString(string(buf)), nil
		}
		xs := make([]Value, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			val, err := convertReflectValue(rv.Index(i))
			if err != nil {
				return Value{}, err
			}
			xs[i] = val
		}
		return VArray(xs), nil

	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return Value{}, fmt.Errorf("cannot convert map with key type %s; only string keys are supported", rv.Type().Key())
		}
		m := make(map[string]Value, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			valRV := iter.Value()
			if valRV.Type() == valueType {
				m[key] = valRV.Interface().(Value)
				continue
			}
			val, err := convertReflectValue(valRV)
			if err != nil {
				return Value{}, fmt.Errorf("map value for key %q: %w", key, err)
			}
			m[key] = val
		}
		return VObject(m), nil

	case reflect.Struct:
		rt := rv.Type()
		m := make(map[string]Value, rv.NumField())
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			if field.PkgPath != "" {
				continue // skip unexported
			}
			name := field.Name
			if tag := field.Tag.Get("script"); tag != "" {
				if tag == "-" {
					continue
				}
				name = tag
			}
			val, err := convertReflectValue(rv.Field(i))
			if err != nil {
				return Value{}, fmt.Errorf("field %s: %w", field.Name, err)
			}
			m[name] = val
		}
		return VObject(m), nil

	default:
		if rv.Type() == valueType {
			return rv.Interface().(Value), nil
		}
	}

	return Value{}, fmt.Errorf("unsupported Go value of kind %s (%s)", rv.Kind(), rv.Type())
}

func (v Value) ToGo(out any) error {
	if out == nil {
		return fmt.Errorf("ToGo requires a non-nil destination")
	}
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("ToGo destination must be a non-nil pointer")
	}
	return assignValueToGo(rv.Elem(), v)
}

func assignValueToGo(dst reflect.Value, val Value) error {
	if !dst.CanSet() {
		return fmt.Errorf("cannot set destination %s", dst.Type())
	}
	switch dst.Kind() {
	case reflect.Interface:
		converted, err := valueToInterface(val)
		if err != nil {
			return err
		}
		if converted == nil {
			dst.Set(reflect.Zero(dst.Type()))
		} else {
			dst.Set(reflect.ValueOf(converted))
		}
		return nil
	case reflect.Pointer:
		if val.Kind == NullKind {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return assignValueToGo(dst.Elem(), val)
	case reflect.Struct:
		if val.Kind != ObjectKind {
			return fmt.Errorf("cannot assign %s to struct %s", kindName(val.Kind), dst.Type())
		}
		typeOfDst := dst.Type()
		for i := 0; i < dst.NumField(); i++ {
			field := typeOfDst.Field(i)
			if field.PkgPath != "" {
				continue
			}
			name := field.Name
			if tag := field.Tag.Get("script"); tag != "" {
				if tag == "-" {
					continue
				}
				name = tag
			}
			fv, ok := val.Obj[name]
			if !ok {
				continue
			}
			if err := assignValueToGo(dst.Field(i), fv); err != nil {
				return fmt.Errorf("field %s: %w", field.Name, err)
			}
		}
		return nil
	case reflect.Map:
		if val.Kind == NullKind {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		if val.Kind != ObjectKind {
			return fmt.Errorf("cannot assign %s to map %s", kindName(val.Kind), dst.Type())
		}
		if dst.IsNil() {
			dst.Set(reflect.MakeMapWithSize(dst.Type(), len(val.Obj)))
		} else {
			dst.Set(reflect.MakeMapWithSize(dst.Type(), len(val.Obj)))
		}
		keyType := dst.Type().Key()
		if keyType.Kind() != reflect.String {
			return fmt.Errorf("cannot assign object to map with non-string key %s", keyType)
		}
		for k, elem := range val.Obj {
			keyVal := reflect.ValueOf(k)
			if !keyVal.Type().ConvertibleTo(keyType) {
				return fmt.Errorf("key %q not convertible to %s", k, keyType)
			}
			valSlot := reflect.New(dst.Type().Elem()).Elem()
			if err := assignValueToGo(valSlot, elem); err != nil {
				return fmt.Errorf("map value for key %q: %w", k, err)
			}
			dst.SetMapIndex(keyVal.Convert(keyType), valSlot)
		}
		return nil
	case reflect.Slice:
		if val.Kind == NullKind {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		if val.Kind != ArrayKind {
			return fmt.Errorf("cannot assign %s to slice %s", kindName(val.Kind), dst.Type())
		}
		slice := reflect.MakeSlice(dst.Type(), len(val.Arr), len(val.Arr))
		for i, elem := range val.Arr {
			if err := assignValueToGo(slice.Index(i), elem); err != nil {
				return fmt.Errorf("slice index %d: %w", i, err)
			}
		}
		dst.Set(slice)
		return nil
	case reflect.Array:
		if val.Kind != ArrayKind {
			return fmt.Errorf("cannot assign %s to array %s", kindName(val.Kind), dst.Type())
		}
		if len(val.Arr) != dst.Len() {
			return fmt.Errorf("array length mismatch: have %d need %d", len(val.Arr), dst.Len())
		}
		for i := 0; i < dst.Len(); i++ {
			if err := assignValueToGo(dst.Index(i), val.Arr[i]); err != nil {
				return fmt.Errorf("array index %d: %w", i, err)
			}
		}
		return nil
	case reflect.Bool:
		dst.SetBool(val.AsBool())
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iv := val.AsInt()
		if dst.OverflowInt(iv) {
			return fmt.Errorf("value %d overflows %s", iv, dst.Type())
		}
		dst.SetInt(iv)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		iv := val.AsInt()
		if iv < 0 {
			return fmt.Errorf("cannot assign negative value %d to unsigned %s", iv, dst.Type())
		}
		u := uint64(iv)
		if dst.OverflowUint(u) {
			return fmt.Errorf("value %d overflows %s", iv, dst.Type())
		}
		dst.SetUint(u)
		return nil
	case reflect.Float32, reflect.Float64:
		fv := val.AsFloat()
		if dst.OverflowFloat(fv) {
			return fmt.Errorf("value %g overflows %s", fv, dst.Type())
		}
		dst.SetFloat(fv)
		return nil
	case reflect.String:
		if val.Kind != StringKind {
			dst.SetString(val.String())
		} else {
			dst.SetString(val.S)
		}
		return nil
	default:
		if dst.Type() == valueType {
			dst.Set(reflect.ValueOf(val))
			return nil
		}
		return fmt.Errorf("unsupported destination type %s", dst.Type())
	}
}

func valueToInterface(val Value) (any, error) {
	switch val.Kind {
	case NullKind:
		return nil, nil
	case IntKind:
		return val.I, nil
	case FloatKind:
		return val.F, nil
	case BoolKind:
		return val.B, nil
	case StringKind:
		return val.S, nil
	case ArrayKind:
		res := make([]any, len(val.Arr))
		for i, elem := range val.Arr {
			conv, err := valueToInterface(elem)
			if err != nil {
				return nil, fmt.Errorf("array index %d: %w", i, err)
			}
			res[i] = conv
		}
		return res, nil
	case ObjectKind:
		res := make(map[string]any, len(val.Obj))
		for k, elem := range val.Obj {
			conv, err := valueToInterface(elem)
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", k, err)
			}
			res[k] = conv
		}
		return res, nil
	default:
		return nil, fmt.Errorf("unsupported value kind %s", kindName(val.Kind))
	}
}
