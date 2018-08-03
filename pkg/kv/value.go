package kv

import (
	"fmt"
	"encoding/json"
	"strconv"
	"gopkg.in/yaml.v2"
)

type ValueType int

const (
	StringType         ValueType = iota
	ReferenceType
	NumericType
	ObjectType
	ArrayType
	FallbackType
	BoolType
)

func (vt *ValueType) MarshalJSON() ([]byte, error) {
	switch *vt {
	case StringType:
		return []byte(`"string"`), nil
	case ReferenceType:
		return []byte(`"reference"`), nil
	case NumericType:
		return []byte(`"number"`), nil
	case ObjectType:
		return []byte(`"object"`), nil
	case ArrayType:
		return []byte(`"array"`), nil
	case FallbackType:
		return []byte(`"fallback"`), nil
	case BoolType:
		return []byte(`"boolean"`), nil
	}
	return nil, fmt.Errorf("unknown value type %v", *vt)
}

func (vt *ValueType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, s)
	if err != nil {
		return err
	}
	switch s {
	case "string": *vt = StringType
	case "reference": *vt = ReferenceType
	case "number": *vt = NumericType
	case "object": *vt = ObjectType
	case "array": *vt = ArrayType
	case "fallback": *vt = FallbackType
	case "boolean": *vt = BoolType
	default: return fmt.Errorf("invalid value type to decode: %v", s)
	}
	return nil
}

// refer to json.Number
// A Number represents a JSON number literal.
type Number string

// create new number from int64
func Int64Number(value int64) *Number {
	n := Number(strconv.FormatInt(value, 10))
	return &n
}

// create new number from float64
func Float64Number(value float64) *Number {
	n := Number(strconv.FormatFloat(value, 'f', -1, 64))
	return &n
}

func FromJsonNumber(value json.Number) *Number {
	n := Number(string(value))
	return &n
}

func ToJsonNumber(value Number) json.Number {
	return json.Number(string(value))
}

// String returns the literal text of the number.
func (n *Number) String() string { return string(*n) }

// Float64 returns the number as a float64.
func (n *Number) Float64() (float64, error) {
	return strconv.ParseFloat(string(*n), 64)
}

// Int64 returns the number as an int64.
func (n *Number) Int64() (int64, error) {
	return strconv.ParseInt(string(*n), 10, 64)
}

func (n *Number) Clone() *Number {
	s := *n
	return &s
}

// IsValidNumber reports whether s is a valid JSON number literal.
func IsValidNumber(s string) bool {
	// This function implements the JSON numbers grammar.
	// See https://tools.ietf.org/html/rfc7159#section-6
	// and http://json.org/number.gif

	if s == "" {
		return false
	}

	// Optional -
	if s[0] == '-' {
		s = s[1:]
		if s == "" {
			return false
		}
	}

	// Digits
	switch {
	default:
		return false

	case s[0] == '0':
		s = s[1:]

	case '1' <= s[0] && s[0] <= '9':
		s = s[1:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// . followed by 1 or more digits.
	if len(s) >= 2 && s[0] == '.' && '0' <= s[1] && s[1] <= '9' {
		s = s[2:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// e or E followed by an optional - or + and
	// 1 or more digits.
	if len(s) >= 2 && (s[0] == 'e' || s[0] == 'E') {
		s = s[1:]
		if s[0] == '+' || s[0] == '-' {
			s = s[1:]
			if s == "" {
				return false
			}
		}
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// Make sure we are at the end.
	return s == ""
}

func (n Number) MarshalJSON() ([]byte, error) {
	return []byte(n), nil
}

type Value struct {
	Type     ValueType
	RefValue interface{}
}

// clone a new value and return the pointer
func (v *Value) Clone() *Value {
	var ref interface{}
	switch v.Type {
	case ObjectType:
		ref = v.RefValue.(*ConfigObject).Clone()
	case ArrayType:
		ref = v.RefValue.(*ConfigArray).Clone()
	case ReferenceType:
		ref = v.RefValue.(*ConfigReference).Clone()
	case FallbackType:
		ref = v.RefValue.(*ConfigFallback).Clone()
	case NumericType:
		ref = v.RefValue.(*Number).Clone()
	case StringType:
		ref = v.RefValue
	case BoolType:
		ref = v.RefValue
	}
	return &Value{
		Type: v.Type,
		RefValue: ref,
	}
}

// unwrap and return the inner value
// for object type, key ordering will be lost
func (v *Value) Unwrap() interface{} {
	switch v.Type {
	case ObjectType:
		co := v.RefValue.(*ConfigObject)
		m := make(map[string]interface{})
		for _, k := range co.m.Keys() {
			v := co.m.Get(k)
			m[k] = v.Unwrap()
		}
		return m
	case ArrayType:
		ca := v.RefValue.(*ConfigArray)
		arr := make([]interface{}, 0, len(ca.arr))
		for _, elem := range ca.arr {
			arr = append(arr, elem.Unwrap())
		}
		return arr
	case ReferenceType:
		cr := v.RefValue.(*ConfigReference)
		m := make(map[string]interface{})
		m["labels"]= cr.Labels
		m["path"] = cr.Path
		return m
	case FallbackType:
		// should not happen
		m := make(map[string]interface{})
		return m
	case NumericType:
		f, _ := v.RefValue.(*Number).Float64()
		return f
	case StringType:
		return v.RefValue
	case BoolType:
		return v.RefValue
	}
	return nil
}

func (v *Value) UnwrapPreserveOrder() interface{} {
	switch v.Type {
	case ObjectType:
		co := v.RefValue.(*ConfigObject)

		ms := &MapSlice{}
		for _, k := range co.m.Keys() {
			v := co.m.Get(k)
			item := MapItem{Key: k, Value: v.UnwrapPreserveOrder()}
			*ms = append(*ms, yaml.MapItem(item))
		}
		return ms
	case ArrayType:
		ca := v.RefValue.(*ConfigArray)
		arr := make([]interface{}, 0, len(ca.arr))
		for _, elem := range ca.arr {
			arr = append(arr, elem.Unwrap())
		}
		return arr
	case ReferenceType:
		cr := v.RefValue.(*ConfigReference)
		m := make(map[string]interface{})
		m["labels"]= cr.Labels
		m["path"] = cr.Path
		return m
	case FallbackType:
		// should not happen
		m := make(map[string]interface{})
		return m
	case NumericType:
		f, _ := v.RefValue.(*Number).Float64()
		return f
	case StringType:
		return v.RefValue
	case BoolType:
		return v.RefValue
	}
	return nil
}


func MakeStringValue(src string) *Value {
	return &Value{
		Type:     StringType,
		RefValue: src,
	}
}

func makeNumericValue(src *Number) *Value {
	return &Value{
		Type:     NumericType,
		RefValue: src,
	}
}

func MakeIntValue(src int64) *Value {
	return &Value{
		Type: NumericType,
		RefValue: Int64Number(src),
	}
}

func MakeFloatValue(src float64) *Value {
	return &Value{
		Type: NumericType,
		RefValue: Float64Number(src),
	}
}


func MakeObjectValue(src *ConfigObject) *Value {
	return &Value{
		Type:     ObjectType,
		RefValue: src,
	}
}

func MakeArrayValue(src *ConfigArray) *Value {
	return &Value{
		Type:     ArrayType,
		RefValue: src,
	}
}

func MakeReferenceValue(src *ConfigReference) *Value {
	return &Value{
		Type:     ReferenceType,
		RefValue: src,
	}
}

// fallback value will not be used
func makeFallbackValue(src *ConfigFallback) *Value {
	return &Value{
		Type:     FallbackType,
		RefValue: src,
	}
}

func MakeBoolValue(src bool) *Value {
	return &Value{
		Type:     BoolType,
		RefValue: src,
	}
}
