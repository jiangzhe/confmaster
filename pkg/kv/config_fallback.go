package kv

// delay fallback is because fallback with reference,
// which can not be merged without resolver
type ConfigFallback struct {
	current ConfigInterface
	fallback ConfigInterface
}

func NewConfigFallback(current ConfigInterface, fallback ConfigInterface) *ConfigFallback {
	return &ConfigFallback {
		current: current,
		fallback: fallback,
	}
}

func (cf *ConfigFallback) GetValue(path string) *Value {
	if v := cf.current.GetValue(path); v != nil {
		return v
	}
	if v := cf.fallback.GetValue(path); v != nil {
		return v
	}
	return nil
}

func (cf *ConfigFallback) GetString(path string) string {
	if v := cf.GetValue(path); v != nil {
		switch v.Type {
		case StringType:
			return v.RefValue.(string)
		}
	}
	return ""
}

func (cf *ConfigFallback) GetNumber(path string) *Number {
	if v := cf.GetValue(path); v != nil {
		switch v.Type {
		case NumericType:
			return v.RefValue.(*Number)
		}
	}
	return nil
}

func (cf *ConfigFallback) GetObject(path string) *ConfigObject {
	if v := cf.GetValue(path); v != nil {
		switch v.Type {
		case ObjectType:
			return v.RefValue.(*ConfigObject)
		}
	}
	return nil
}

func (cf *ConfigFallback) GetArray(path string) *ConfigArray {
	if v := cf.GetValue(path); v != nil {
		switch v.Type {
		case ArrayType:
			return v.RefValue.(*ConfigArray)
		}
	}
	return nil
}

func (cf *ConfigFallback) GetReference(path string) *ConfigReference {
	if v := cf.GetValue(path); v != nil {
		switch v.Type {
		case ReferenceType:
			return v.RefValue.(*ConfigReference)
		}
	}
	return nil
}

func (cf *ConfigFallback) Refs() map[string]*ConfigReference {
	refs := make(map[string]*ConfigReference)
	for k, v := range cf.fallback.Refs() {
		refs[k] = v
	}
	for k, v := range cf.current.Refs() {
		refs[k] = v
	}
	return refs
}

func (cf *ConfigFallback) Keys() []string {
	km := make(map[string]bool)
	keys := make([]string, 0)
	for _, k := range cf.fallback.Keys() {
		if !km[k] {
			km[k] = true
			keys = append(keys, k)
		}
	}
	for _, k := range cf.current.Keys() {
		if !km[k] {
			km[k] = true
			keys = append(keys, k)
		}
	}
	return keys
}

func (cf *ConfigFallback) WithFallback(fallback ConfigInterface) ConfigInterface {
	switch fallback.(type) {
	case *ConfigObject:
		return cf.withFallbackObject(fallback.(*ConfigObject))
	}
	return cf
}

func (cf *ConfigFallback) Clone() *ConfigFallback {
	var current ConfigInterface
	var fallback ConfigInterface
	switch cf.current.(type) {
	case *ConfigObject:
		current = cf.current.(*ConfigObject).Clone()
	case *ConfigFallback:
		current = cf.current.(*ConfigFallback).Clone()
	}
	switch cf.fallback.(type) {
	case *ConfigObject:
		fallback = cf.fallback.(*ConfigObject).Clone()
	case *ConfigFallback:
		fallback = cf.fallback.(*ConfigFallback).Clone()
	}
	return &ConfigFallback{
		current: current,
		fallback: fallback,
	}
}

func (cf *ConfigFallback) withFallbackObject(fallback *ConfigObject) ConfigInterface {
	switch cf.fallback.(type) {
	case *ConfigObject:
		fb := cf.fallback.(*ConfigObject).WithFallback(fallback)
		return NewConfigFallback(cf.current, fb)
	case *ConfigFallback:
		fb := cf.fallback.(*ConfigFallback).WithFallback(fallback)
		return NewConfigFallback(cf.current, fb)
	}

	return NewConfigFallback(cf, fallback)
}
