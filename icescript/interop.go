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
