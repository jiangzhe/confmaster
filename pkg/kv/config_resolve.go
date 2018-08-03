package kv

import "gopkg.in/yaml.v2"

type Resolver interface {
	// return a resolved config, all inner references will be resolved after this call
	// error out if any reference is not resolvable
	Resolve(ConfigInterface) (ResolvedConfigInterface, error)
}

// special type to retain key order of map
type MapSlice yaml.MapSlice

type MapItem yaml.MapItem

type ResolvedConfigInterface interface {

	ToConfig() *ConfigObject

	ToMap() map[string]interface{}

	ToMapSlice() MapSlice

	Format(formatter Formatter) ([]byte, error)
}

// any config object without references can be treated as
// a ResolvedConfigObject
type ResolvedConfigObject struct {
	m *LinkedMap
}

func (rco *ResolvedConfigObject) ToConfig() *ConfigObject {
	return &ConfigObject{
		m: rco.m,
	}
}

func (rco *ResolvedConfigObject) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	for _, k := range rco.m.Keys() {
		v := rco.m.Get(k)
		m[k] = v.Unwrap()
	}
	return m
}

func (rco *ResolvedConfigObject) ToMapSlice() MapSlice {
	ms := MapSlice{}
	for _, k := range rco.m.Keys() {
		v := rco.m.Get(k)
		item := yaml.MapItem{
			Key: k,
			Value: v.UnwrapPreserveOrder(),
		}
		ms = append(ms, item)
	}
	return ms
}

func (rco *ResolvedConfigObject) Format(formatter Formatter) ([]byte, error) {
	return formatter.Format(rco)
}
