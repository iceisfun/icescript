package object

import (
	"testing"
)

func TestObjectConversions(t *testing.T) {
	tests := []struct {
		name     string
		obj      Object
		expected struct {
			asFloat  float64
			asInt    int64
			asString string
			asBool   bool
		}
		ok struct {
			asFloat  bool
			asInt    bool
			asString bool
			asBool   bool
		}
	}{
		{
			name: "Integer(5)",
			obj:  &Integer{Value: 5},
			expected: struct {
				asFloat  float64
				asInt    int64
				asString string
				asBool   bool
			}{5.0, 5, "5", true},
			ok: struct {
				asFloat  bool
				asInt    bool
				asString bool
				asBool   bool
			}{true, true, true, true},
		},
		{
			name: "Float(3.14)",
			obj:  &Float{Value: 3.14},
			expected: struct {
				asFloat  float64
				asInt    int64
				asString string
				asBool   bool
			}{3.14, 3, "3.140000", true},
			ok: struct {
				asFloat  bool
				asInt    bool
				asString bool
				asBool   bool
			}{true, true, true, true},
		},
		{
			name: "Boolean(true)",
			obj:  &Boolean{Value: true},
			expected: struct {
				asFloat  float64
				asInt    int64
				asString string
				asBool   bool
			}{1.0, 1, "true", true},
			ok: struct {
				asFloat  bool
				asInt    bool
				asString bool
				asBool   bool
			}{true, true, true, true},
		},
		{
			name: "Boolean(false)",
			obj:  &Boolean{Value: false},
			expected: struct {
				asFloat  float64
				asInt    int64
				asString string
				asBool   bool
			}{0.0, 0, "false", false},
			ok: struct {
				asFloat  bool
				asInt    bool
				asString bool
				asBool   bool
			}{true, true, true, true},
		},
		{
			name: "String(\"10\")",
			obj:  &String{Value: "10"},
			expected: struct {
				asFloat  float64
				asInt    int64
				asString string
				asBool   bool
			}{10.0, 10, "10", false},
			ok: struct {
				asFloat  bool
				asInt    bool
				asString bool
				asBool   bool
			}{true, true, true, true},
		},
		{
			name: "String(\"true\")",
			obj:  &String{Value: "true"},
			expected: struct {
				asFloat  float64
				asInt    int64
				asString string
				asBool   bool
			}{0.0, 0, "true", true},
			ok: struct {
				asFloat  bool
				asInt    bool
				asString bool
				asBool   bool
			}{false, false, true, true},
		},
		{
			name: "Null",
			obj:  &Null{},
			expected: struct {
				asFloat  float64
				asInt    int64
				asString string
				asBool   bool
			}{0, 0, "", false},
			ok: struct {
				asFloat  bool
				asInt    bool
				asString bool
				asBool   bool
			}{false, false, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// AsFloat
			fVal, fOk := tt.obj.AsFloat()
			if fOk != tt.ok.asFloat {
				t.Errorf("AsFloat ok = %v, want %v", fOk, tt.ok.asFloat)
			}
			if fOk && fVal != tt.expected.asFloat {
				t.Errorf("AsFloat value = %v, want %v", fVal, tt.expected.asFloat)
			}

			// AsInt
			iVal, iOk := tt.obj.AsInt()
			if iOk != tt.ok.asInt {
				t.Errorf("AsInt ok = %v, want %v", iOk, tt.ok.asInt)
			}
			if iOk && iVal != tt.expected.asInt {
				t.Errorf("AsInt value = %v, want %v", iVal, tt.expected.asInt)
			}

			// AsString
			sVal, sOk := tt.obj.AsString()
			if sOk != tt.ok.asString {
				t.Errorf("AsString ok = %v, want %v", sOk, tt.ok.asString)
			}
			if sOk && sVal != tt.expected.asString {
				t.Errorf("AsString value = %v, want %v", sVal, tt.expected.asString)
			}

			// AsBool
			bVal, bOk := tt.obj.AsBool()
			if bOk != tt.ok.asBool {
				t.Errorf("AsBool ok = %v, want %v", bOk, tt.ok.asBool)
			}
			if bOk && bVal != tt.expected.asBool {
				t.Errorf("AsBool value = %v, want %v", bVal, tt.expected.asBool)
			}
		})
	}
}
