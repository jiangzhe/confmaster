package kv

import (
	"errors"
)

type ConfigInterface interface {
	GetValue(path string) *Value
	GetString(path string) string
	GetInt(path string) int
	GetObject(path string) *ConfigObject
	GetArray(path string) *ConfigArray
	GetReference(path string) *ConfigReference
	Refs() []string
	Keys() []string
	WithFallback(fallback ConfigInterface) ConfigInterface
}

type ResolvedConfigInterface interface {
	ConfigInterface

	ToMap() map[string]interface{}
}

var (
	ErrConfigNotExists = errors.New("config not exists")
	ErrConfigChangeNotAllowed = errors.New("config change not allowed")
	ErrConfigCastInvalid = errors.New("config cast invalid")
	ErrValueReferenceConflict = errors.New("value reference conflict")
	ErrValuePathConflict = errors.New("value path conflict")
)

type Resolver interface {
	// return a resolved config, all inner references will be resolved after this call
	// error out if any reference is not resolvable
	Resolve(ConfigInterface) (ResolvedConfigInterface, error)
}
