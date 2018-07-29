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

func (cf *ConfigFallback) GetInt(path string) int {
	if v := cf.GetValue(path); v != nil {
		switch v.Type {
		case NumericType:
			return v.RefValue.(int)
		}
	}
	return 0
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

func (cf *ConfigFallback) Refs() []string {
	rm := make(map[string]bool)
	refs := make([]string, 0)
	for _, r := range cf.fallback.Refs() {
		rm[r] = true
	}
	for _, r := range cf.current.Refs() {
		rm[r] = true
	}
	for k := range rm {
		refs = append(refs, k)
	}
	return refs
}

func (cf *ConfigFallback) Keys() []string {
	km := make(map[string]bool)
	keys := make([]string, 0)
	for _, k := range cf.fallback.Keys() {
		km[k] = true
	}
	for _, k := range cf.current.Keys() {
		km[k] = true
	}
	for k := range km {
		keys = append(keys, k)
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
