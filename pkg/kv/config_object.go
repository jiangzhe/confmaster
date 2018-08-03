package kv

import (
	"strings"
	"fmt"
)

type ConfigObject struct {
	m *LinkedMap
}

// merge given config object to current one
func (co *ConfigObject) merge(newco *ConfigObject) error {
	if newco.m.Len() == 0 {
		return nil
	}
	for _, k := range newco.m.Keys() {
		v := newco.m.Get(k)
		if err := co.setKeyValue(k, v); err != nil {
			return err
		}
	}
	return nil
}

// set key value which key does not contain dot
func (co *ConfigObject) setKeyValue(key string, value *Value) error {
	key, idx, err := parseArrayKey(key)
	if err != nil {
		return err
	}
	if idx == 0 {
		// try create array if non-match
		v := co.m.Get(key)
		if v == nil || v.Type != ArrayType {
			// create a new array
			v = MakeArrayValue(NewConfigArray())
			ca := v.RefValue.(*ConfigArray)
			ca.arr = append(ca.arr, value)
			co.m.Put(key, v)
			return nil
		}
		ca := v.RefValue.(*ConfigArray)
		return ca.setValue(idx, value)
	} else if idx > 0 {
		v := co.m.Get(key)
		if v == nil || v.Type != ArrayType {
			return fmt.Errorf("value path is invalid: %v", key)
		}
		ca := v.RefValue.(*ConfigArray)
		return ca.setValue(idx, value)
	}
	v := co.m.Get(key)
	if v == nil {
		co.m.Put(key, value)
		return nil
	}
	if v.Type == ObjectType && value.Type == ObjectType {
		origco := v.RefValue.(*ConfigObject)
		currco := value.RefValue.(*ConfigObject)
		if err := origco.merge(currco); err != nil {
			return err
		}
		return nil
	}
	co.m.Put(key, value)
	return nil
}


func (co *ConfigObject) setValue(path string, value *Value) error {
	if !strings.Contains(path, ".") {
		return co.setKeyValue(path, value)
	}
	paths := strings.Split(path, ".")

	var key string
	var obj = co
	for i := 0; i < len(paths)-1; i++ {
		key = paths[i]
		// auto merge if target is also object
		v := obj.getValueByKey(key)
		if v == nil || v.Type != ObjectType {
			v = MakeObjectValue(NewConfigObject())
			obj.setKeyValue(key, v)
		}
		obj = v.RefValue.(*ConfigObject)
	}
	key = paths[len(paths)-1]
	return obj.setKeyValue(key, value)
}

// implements valueSetter

func (co *ConfigObject) setString(path string, value string) error {
	return co.setValue(path, MakeStringValue(value))
}

func (co *ConfigObject) setNumber(path string, value *Number) error {
	return co.setValue(path, makeNumericValue(value))
}


func (co *ConfigObject) setObject(path string, value *ConfigObject) error {
	return co.setValue(path, MakeObjectValue(value))
}

func (co *ConfigObject) setArray(path string, value *ConfigArray) error {
	return co.setValue(path, MakeArrayValue(value))
}

func (co *ConfigObject) setReference(path string, value *ConfigReference) error {
	return co.setValue(path, MakeReferenceValue(value))
}

func (co *ConfigObject) getValueByKey(key string) *Value {
	key, idx, err := parseArrayKey(key)
	if err != nil {
		return nil
	}
	if idx == -1 {
		return co.m.Get(key)
	}
	v := co.m.Get(key)
	if v == nil || v.Type != ArrayType {
		return nil
	}
	ca := v.RefValue.(*ConfigArray)
	return ca.getValue(idx)
}

func (co *ConfigObject) traverse(f traverseFunc) error {
	var err error
	for _, k := range co.m.Keys() {
		v := co.m.Get(k)
		if err = traverse(k, v, f); err != nil {
			return err
		}
	}
	return err
}

// implements ConfigInterface

func (co *ConfigObject) GetValue(path string) *Value {
	if !strings.Contains(path, ".") {
		return co.getValueByKey(path)
	}

	paths := strings.Split(path, ".")
	var value *Value
	var obj = co
	for _, k := range paths[:len(paths)-1] {
		value = obj.getValueByKey(k)
		if value == nil || value.Type != ObjectType {
			 return nil
		}
		obj = value.RefValue.(*ConfigObject)
	}
	key := paths[len(paths)-1]
	return obj.getValueByKey(key)
}

func (co *ConfigObject) GetString(path string, defaultValue string) string {
	if value := co.GetValue(path); value != nil {
		switch value.Type {
		case StringType: return value.RefValue.(string)
		}
	}
	return defaultValue
}

func (co *ConfigObject) GetBool(path string, defaultValue bool) bool {
	if value := co.GetValue(path); value != nil {
		switch value.Type {
		case BoolType: return value.RefValue.(bool)
		}
	}
	return defaultValue
}

func (co *ConfigObject) GetNumber(path string) *Number {
	if value := co.GetValue(path); value != nil {
		switch value.Type {
		case NumericType: return value.RefValue.(*Number)
		}
	}
	return nil
}

func (co *ConfigObject) GetObject(path string) *ConfigObject {
	if value := co.GetValue(path); value != nil {
		switch value.Type {
		case ObjectType: return value.RefValue.(*ConfigObject)
		}
	}
	return nil
}

func (co *ConfigObject) GetArray(path string) *ConfigArray {
	if value := co.GetValue(path); value != nil {
		switch value.Type {
		case ArrayType: return value.RefValue.(*ConfigArray)
		}
	}
	return nil
}

func (co *ConfigObject) GetReference(path string) *ConfigReference {
	if value := co.GetValue(path); value != nil {
		switch value.Type {
		case ReferenceType: return value.RefValue.(*ConfigReference)
		}
	}
	return nil
}

func (co *ConfigObject) Refs() map[string]*ConfigReference {
	refs := make(map[string]*ConfigReference)
	co.traverse(func(path string, value *Value) error {
		switch value.Type {
		case ReferenceType:
			refs[path] = value.RefValue.(*ConfigReference)
		}
		return nil
	})
	return refs
}

func (co *ConfigObject) Keys() []string {
	res := make([]string, co.m.Len())
	copy(res, co.m.Keys())
	return res
}

func (co *ConfigObject) Clone() *ConfigObject {
	m := NewLinkedMap(co.m.Len())
	for _, k := range co.m.Keys() {
		v := co.m.Get(k)
		m.Put(k, v.Clone())
	}
	return &ConfigObject{
		m: m,
	}
}

func (co *ConfigObject) WithFallback(fallback ConfigInterface) ConfigInterface {
	switch fallback.(type) {
	case *ConfigObject:
		return co.withFallbackObject(fallback.(*ConfigObject))
	case *ConfigFallback:
		return NewConfigFallback(co, fallback)
	}
	return co
}

func (co *ConfigObject) withFallbackObject(fallback *ConfigObject) ConfigInterface {
	for path := range fallback.Refs() {
		if v := co.GetValue(path); v != nil {
			// if object, cannot determine which key will be overwrite,
			// so delay fallback
			switch v.Type {
			case ObjectType:
				return NewConfigFallback(co, fallback)
			}
		}
	}
	for path := range co.Refs() {
		if v := fallback.GetValue(path); v != nil {
			// if fallback is object, delay it
			switch v.Type {
			case ObjectType:
				return NewConfigFallback(co, fallback)
			}
		}
	}

	// fallback as base, preserve key order
	result := fallback.Clone()
	for _, k := range co.m.Keys() {
		v := co.m.Get(k)
		result.setValue(k, v.Clone())
	}
	return result
}

func NewConfigObject() *ConfigObject {
	return &ConfigObject{
		m: NewLinkedMap(0),
	}
}


