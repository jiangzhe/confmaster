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

var (
	ErrConfigNotExists = errors.New("config not exists")
	ErrConfigChangeNotAllowed = errors.New("config change not allowed")
	ErrConfigCastInvalid = errors.New("config cast invalid")
	ErrValueReferenceConflict = errors.New("value reference conflict")
	ErrValuePathConflict = errors.New("value path conflict")
)
