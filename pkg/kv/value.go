package kv

import (
	"fmt"
	"encoding/json"
)

type ValueType int

const (
	StringType         ValueType = iota
	ReferenceType
	NumericType
	ObjectType
	ArrayType
	FallbackType
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
	default: return fmt.Errorf("invalid value type to decode: %v", s)
	}
	return nil
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
	case StringType:
		fallthrough
	case NumericType:
		ref = v.RefValue
	}
	return &Value{
		Type: v.Type,
		RefValue: ref,
	}
}

// unwrap and return the inner value
func (v *Value) Unwrap() interface{} {
	switch v.Type {
	case ObjectType:
		co := v.RefValue.(*ConfigObject)
		m := make(map[string]interface{})
		for k, v := range *co.m {
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
	case StringType:
		fallthrough
	case NumericType:
		return v.RefValue
	}
	return nil
}

// value setter is an internal interface to modify the inner value
// it is only for internal usage,
// if user does not call these methods, the ConfigInterface can be
// treated as immutable
type valueSetter interface {
	setString(name string, value string)
	setNumber(name string, value float64)
	setObject(name string, value *ConfigObject)
	setArray(name string, value *ConfigArray)
	setReference(name string, value *ConfigReference)
	setValue(path string, t ValueType, ref interface{})
	unset(path string)
}

func MakeStringValue(src string) *Value {
	return &Value{
		Type:     StringType,
		RefValue: src,
	}
}

func MakeNumericValue(src float64) *Value {
	return &Value{
		Type:     NumericType,
		RefValue: src,
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
