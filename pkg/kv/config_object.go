package kv

import (
	"strings"
	"fmt"
)

type ConfigObject struct {
	m *map[string]*Value
	refs map[string]*ConfigReference
}

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



// merge given config object to current one
func (co *ConfigObject) merge(newco *ConfigObject) error {
	if len(*newco.m) == 0 {
		return nil
	}
	for k, v := range *newco.m {
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
		v := (*co.m)[key]
		if v == nil || v.Type != ArrayType {
			// create a new array
			v = MakeArrayValue(NewConfigArray())
			ca := v.RefValue.(*ConfigArray)
			ca.arr = append(ca.arr, value)
			(*co.m)[key] = v
			return nil
		}
		ca := v.RefValue.(*ConfigArray)
		return ca.setValue(idx, value)
	} else if idx > 0 {
		v := (*co.m)[key]
		if v == nil || v.Type != ArrayType {
			return fmt.Errorf("value path is invalid: %v", key)
		}
		ca := v.RefValue.(*ConfigArray)
		return ca.setValue(idx, value)
	}
	// invalid idx, so use object to set key value
	v := (*co.m)[key]
	if v == nil {
		(*co.m)[key] = value
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
	(*co.m)[key] = value
	return nil
}

func (co *ConfigObject) setString(path string, value string) error {
	return co.setValue(path, MakeStringValue(value))
}

func (co *ConfigObject) setNumber(path string, value *Number) error {
	return co.setValue(path, makeNumericValue(value))
}

// todo: set value
// todo: ref check
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
		return (*co.m)[key]
	}
	// find element in array
	v := (*co.m)[key]
	if v == nil || v.Type != ArrayType {
		return nil
	}
	ca := v.RefValue.(*ConfigArray)
	return ca.getValue(idx)
}

func (co *ConfigObject) traverse(f traverseFunc) {
	if len(*co.m) == 0 {
		return
	}
	
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
		result.setValue(k, v.Clone())
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


