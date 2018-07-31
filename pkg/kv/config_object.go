package kv

import (
	"strings"
	"errors"
)

type ConfigObject struct {
	m *map[string]*Value
	refs map[string]*ConfigReference
}


var (
	ErrIndexOutOfBound = errors.New("key index out of bound")
)

// given a path, traverse and find the key until last path,
// immediately return nil if key is not found
func (co *ConfigObject) findKey(path string) (*ConfigObject, string) {
	obj := co
	paths := strings.Split(path, ".")
	if len(paths) == 1 {
		return obj, path
	}
	for _, p := range paths[:len(paths)-1] {
		v, _ := (*obj.m)[p]
		if v == nil || v.Type != ObjectType {
			return nil, p
		}
	}
	return obj, paths[len(paths)-1]
}

// todo
// recursively add ref
func (co *ConfigObject) addRef(path string, ref *ConfigReference) {
	prefixIdx := strings.Index(path, ".")
	if prefixIdx != -1 {
		if so := co.GetObject(path[:prefixIdx]); so != nil {
			so.addRef(path[prefixIdx+1:], ref)
		}
	}
	co.refs[path] = ref
}

// todo
// recursively remove ref
func (co *ConfigObject) removeRef(path string) {
	prefixIdx := strings.Index(path, ".")
	if prefixIdx != -1 {
		if so := co.GetObject(path[:prefixIdx]); so != nil {
			so.removeRef(path[prefixIdx+1:])
		}
	}
	delete(co.refs, path)
}

func (co *ConfigObject) setString(path string, value string) error {
	return setValue(co, path, MakeStringValue(value))
}

func (co *ConfigObject) setNumber(path string, value *Number) error {
	return setValue(co, path, makeNumericValue(value))
}

// todo
// might have ref
func (co *ConfigObject) setObject(path string, value *ConfigObject) error {
	if obj := co.GetValue(path); obj != nil && obj.Type == ObjectType {
		orig := obj.RefValue.(*ConfigObject)
		for k, v := range *value.m {
			if obj, ok := (*orig.m)[k]; ok {
				if obj.Type == v.Type {
					switch obj.Type {
					case ObjectType:
						co.setObject(path + "." + k, v.RefValue.(*ConfigObject))
					case ArrayType:
						array := obj.RefValue.(*ConfigArray)
						array.arr = append(array.arr, v.RefValue.(*ConfigArray).arr...)
					case ReferenceType:
						fallthrough
					case StringType:
						fallthrough
					case NumericType:
						fallthrough
					case FallbackType:
						co.setValue(path + "." + k, v.Type, v.RefValue)
					}
				} else if err := co.setValue(path + "." + k, ObjectType, value); err != nil {
					return err
				}
			} else if err := co.setValue(path + "." + k, ObjectType, value); err != nil {
				return err
			}
		}
		return nil
	} else {
		return co.setValue(path, ObjectType, value)
	}
}

func (co *ConfigObject) setArray(path string, value *ConfigArray) error {
	return setValue(co, path, MakeArrayValue(value))
}

func (co *ConfigObject) setReference(path string, value *ConfigReference) error {
	return setValue(co, path, MakeReferenceValue(value))
}

func (co *ConfigObject) getValueByKey(key string) *Value {
	return (*co.m)[key]
}

// todo
func (co *ConfigObject) unset(path string) error {
	obj, key := co.findKey(path)
	if obj != nil {
		delete(*obj.m, key)
	}
	return nil
}

func (co *ConfigObject) GetValue(path string) (res *Value) {
	if obj, key := traversePath(co, path); obj != nil {
		if v, ok := (*obj.m)[key]; ok {
			res = v
		}
	}
	return res
}

func (co *ConfigObject) GetString(path string) (res string) {
	if obj, key := traversePath(co, path); obj != nil {
		if v, ok := (*obj.m)[key]; ok {
			switch v.Type {
			case StringType:
				res = v.RefValue.(string)
			}
		}
	}
	return
}

func (co *ConfigObject) GetNumber(path string) (res *Number) {
	if obj, key := traversePath(co, path); obj != nil {
		if v, ok := (*obj.m)[key]; ok {
			switch v.Type {
			case NumericType:
				res = v.RefValue.(*Number)
			}
		}
	}
	return res
}

func (co *ConfigObject) GetObject(path string) (res *ConfigObject) {
	if obj, key := traversePath(co, path); obj != nil {
		if v, ok := (*obj.m)[key]; ok {
			switch v.Type {
			case ObjectType:
				res = v.RefValue.(*ConfigObject)
			}
		}
	}
	return res
}

func (co *ConfigObject) GetArray(path string) (res *ConfigArray) {
	if obj, key := traversePath(co, path); obj != nil {
		if v, ok := (*obj.m)[key]; ok {
			switch v.Type {
			case ArrayType:
				res = v.RefValue.(*ConfigArray)
			}
		}
	}
	return res
}

func (co *ConfigObject) GetReference(path string) (res *ConfigReference) {
	if obj, key := traversePath(co, path); obj != nil {
		if v, ok := (*obj.m)[key]; ok {
			switch v.Type {
			case ReferenceType:
				res = v.RefValue.(*ConfigReference)
			}
		}
	}
	return res
}

func (co *ConfigObject) Refs() []string {
	refs := make([]string, len(co.refs))
	i := 0
	for r := range co.refs {
		refs[i] = r
		i++
	}
	return refs
}

func (co *ConfigObject) Keys() []string {
	res := make([]string, len(*co.m))
	i := 0
	for k := range *co.m {
		res[i] = k
		i++
	}
	return res
}

func (co *ConfigObject) Clone() *ConfigObject {
	m := make(map[string]*Value, len(*co.m))
	for k, v := range *co.m {
		m[k] = v.Clone()
	}
	refs := make(map[string]*ConfigReference, len(co.refs))
	for k, v := range co.refs {
		refs[k] = v.Clone()
	}
	return &ConfigObject{
		m: &m,
		refs: refs,
	}
}

func (co *ConfigObject) WithFallback(fallback ConfigInterface) ConfigInterface {
	switch fallback.(type) {
	case *ConfigObject:
		return co.withFallbackObject(fallback.(*ConfigObject))
	}
	return co
}

func (co *ConfigObject) withFallbackObject(fallback *ConfigObject) ConfigInterface {
	for _, r := range fallback.Refs() {
		if v := co.GetValue(r); v != nil {
			// if object, cannot determine which key will be overwrite,
			// so delay fallback
			switch v.Type {
			case ObjectType:
				return NewConfigFallback(co, fallback)
			}
		}
	}
	for _, r := range co.Refs() {
		if v := fallback.GetValue(r); v != nil {
			// if fallback is object, delay it
			switch v.Type {
			case ObjectType:
				return NewConfigFallback(co, fallback)
			}
		}
	}

	result := fallback.Clone()
	for k, v := range *co.m {
		result.setValue(k, v.Type, v.Clone().RefValue)
	}
	return result
}

// traverse the path until the last path,
// return the final object and path
func traversePath(co *ConfigObject, path string) (obj *ConfigObject, key string) {
	paths := strings.Split(path, ".")
	obj = co
	key = path
	if len(paths) == 1 {
		return
	}
	var value *Value
	for _, key = range paths[:len(paths)-1] {
		if value, _ = (*obj.m)[key]; value == nil {
			return nil, key
		}
		switch value.Type {
		case ObjectType:
			obj = value.RefValue.(*ConfigObject)
		default:
			return nil, key
		}
	}
	key = paths[len(paths)-1]
	return
}

func NewConfigObject() *ConfigObject {
	m := make(map[string]*Value)
	co := ConfigObject{
		m: &m,
		refs: make(map[string]*ConfigReference),
	}
	return &co
}


