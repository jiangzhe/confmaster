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
	//setValue(path string, value *Value) error
	unset(path string) error
}

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

// todo: ref check
func setValue(vs valueSetter, path string, value *Value) error {
	if !strings.Contains(path, ".") {
		switch vs.(type) {
		case *ConfigObject:
			co := vs.(*ConfigObject)

			if path == "0" {
				return ErrTopLevelArrayNotAllowed
			}
			(*co.m)[path] = value
			return nil
		case *ConfigArray:
			ca := vs.(*ConfigArray)
			// to set a non-numeric key to array is error
			n, err := strconv.Atoi(path)
			if err != nil {
				return ErrInvalidKeyOnArray
			}
			if n >= 0 && n < len(ca.arr) {
				ca.arr[n] = value
				return nil
			} else if n == len(ca.arr) {
				ca.arr = append(ca.arr, value)
				return nil
			}
			return ErrIndexOutOfBound
		default:
			return ErrInvalidValueSetter
		}
	}
	paths := strings.Split(path, ".")
	var parent = vs
	var pkey = paths[0]
	var currvalue *Value
	var curr valueSetter
	var key string
	//var exists bool

	currvalue = parent.getValueByKey(pkey)
	if currvalue == nil {
		if err := setValue(parent, pkey, MakeObjectValue(NewConfigObject())); err != nil {
			return err
		}
		currvalue = parent.getValueByKey(pkey)
	}
	currvalue, curr = updateValueIfNonmatch(currvalue, &parent, pkey)

pathloop:
	for i := 1; i < len(paths)-1; i++ {
		key = paths[i]
		if n, err := strconv.Atoi(key); err == nil {
			// is a number
			switch currvalue.Type {
			case ArrayType:
				ca := curr.(*ConfigArray)
				if n >= 0 && n < len(ca.arr) {
					// fit an element in array
					parent = curr
					pkey = key
					currvalue = ca.arr[n]
					currvalue, curr = updateValueIfNonmatch(currvalue, &parent, pkey)
					continue pathloop
				} else if n == len(ca.arr) {
					if err := ca.setObject(key, NewConfigObject()); err != nil {
						return err
					}
					parent = curr
					pkey = key
					currvalue = ca.arr[n]
					currvalue, curr = updateValueIfNonmatch(currvalue, &parent, pkey)
					continue pathloop
				} else {
					// is a number but out of array bound
					// just return the error
					return ErrIndexOutOfBound
				}
			}
			// curr is not a array, but key=0, change to array for parent
			if n == 0 {
				parent.setArray(pkey, NewConfigArray())
				currvalue = parent.getValueByKey(pkey)
				curr = currvalue.RefValue.(*ConfigArray)

				parent = curr
				pkey = key
				// append key=0 to curr array
				curr.setObject(key, NewConfigObject())
				currvalue = curr.getValueByKey(key)
				curr = currvalue.RefValue.(*ConfigObject)
				continue pathloop
			}
		}
		// key as a string, convert all non-object to object
		switch currvalue.Type {
		case ObjectType:
			co := curr.(*ConfigObject)
			var exists bool
			currvalue, exists = (*co.m)[key]
			if !exists {
				parent = curr
				pkey = key
				currvalue = MakeObjectValue(NewConfigObject())
				(*co.m)[key] = currvalue
				currvalue, curr = updateValueIfNonmatch(currvalue, &parent, pkey)
				continue pathloop
			}
			parent = curr
			pkey = key
			currvalue, curr = updateValueIfNonmatch(currvalue, &parent, pkey)
		default:
			// non-object type should be modified as object
			parent.setObject(pkey, NewConfigObject())
			currvalue = parent.getValueByKey(pkey)
			co := currvalue.RefValue.(*ConfigObject)

			parent = curr
			pkey = key
			currvalue = MakeObjectValue(NewConfigObject())
			(*co.m)[key] = currvalue
			currvalue, curr = updateValueIfNonmatch(currvalue, &parent, pkey)
		}
	}
	key = paths[len(paths)-1]
	// final setter and key
	return setValue(curr, key, value)
}

// for object and array , just return the same type
// for others, change the type to object and return
func updateValueIfNonmatch(currvalue *Value, parent *valueSetter, pkey string) (*Value, valueSetter) {
	switch currvalue.Type {
	case ObjectType:
		return currvalue, currvalue.RefValue.(*ConfigObject)
	case ArrayType:
		return currvalue, currvalue.RefValue.(*ConfigArray)
	default:
		// todo: need to check if ref is removed
		newco := NewConfigObject()
		(*parent).setObject(pkey, newco)
		newvalue := (*parent).getValueByKey(pkey)
		return newvalue, newvalue.RefValue.(*ConfigObject)
	}
}