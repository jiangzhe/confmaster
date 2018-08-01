package kv

import (
	"strings"
	"strconv"
	"errors"
	"fmt"
)


// value setter is an internal interface to modify the inner value
// it is only for internal usage,
// if user does not call these methods, the ConfigInterface can be
// treated as immutable
type valueSetter interface {
	setString(path string, value string) error
	setNumber(path string, value *Number) error
	setObject(path string, value *ConfigObject) error
	setArray(path string, value *ConfigArray) error
	setReference(path string, value *ConfigReference) error

	getValueByKey(key string) *Value

	traverse(traverseFunc)
}

type traverseFunc func(path string, value *Value)

var (
	ErrTopLevelArrayNotAllowed = errors.New("top level array not allowed")
	ErrInvalidKeyOnArray = errors.New("invalid key on array")
	ErrInvalidValueSetter = errors.New("invalid value setter")
)

// parse key to see if it match array key pattern
// <array-name>[<array-index>]
// array-name should be at least one char
// array-index is an integer
// only one-dimension array is allowed
func parseArrayKey(key string) (string, int, error) {
	if len(key) < 4 {
		return key, -1, nil
	}
	if strings.HasSuffix(key, "]") {
		start := strings.LastIndex(key, "[")
		if start <= 0 {
			return key, -1, nil
		}
		idx := key[start+1:len(key)-1]
		i, err := strconv.Atoi(idx)
		if err != nil {
			return key, -1, nil
		}
		if i < 0 {
			return key, -1, fmt.Errorf("invalid key index: %v", i)
		}
		return key[:start], i, nil
	}
	return key, -1, nil
}

func traverse(path string, value *Value, f traverseFunc) {
	switch value.Type {
	case StringType:
		f(path, value)
	case NumericType:
		f(path, value)
	case ReferenceType:
		f(path, value)
	case ObjectType:
		f(path, value)
		// inner loop
		co := value.RefValue.(*ConfigObject)
		for k, v := range *co.m {
			traverse(path + "." + k, v, f)
		}
	case ArrayType:
		f(path, value)
		// inner loop
		ca := value.RefValue.(*ConfigArray)
		for idx, elem := range ca.arr {
			traverse(fmt.Sprintf("%v[%v]", path, idx), elem, f)
		}
	}
}