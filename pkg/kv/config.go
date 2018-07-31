package kv

import (
	"errors"
	"encoding/json"
	"bytes"
	"go/types"
	"strconv"
	"sort"
	"fmt"
)

type ConfigInterface interface {
	GetValue(path string) *Value
	GetString(path string) string
	GetNumber(path string) *Number
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

// read json bytes and returns config
func ConfigFromJson(input []byte) (ConfigInterface, error) {
	m, err := mapFromJson(input)
	if err != nil {
		return nil, err
	}
	mp := &mapProcessor{
		root: NewConfigObject(),
		m: m,
	}
	if err = mp.traverse(); err != nil {
		return nil, err
	}
	return mp.root, nil
}

func mapFromJson(input []byte) (m map[string]interface{}, err error) {
	buf := bytes.NewBuffer(input)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	m = make(map[string]interface{})
	err = decoder.Decode(&m)
	return
}

// parse config object from map,
// if map is empty, return nil
//
// key rules:
//
// 1. key must be string
// 2. if key has dot inside, it is treated as a path
//    path will be split and the value will be set through it, leading to
//    a nested structure generated
//    e.g. {"a.b.c":"hello"} will be converted to {"a":{"b":{"c":"hello"}}}
// 3. if part of the path is a number-like string, effort will be tried
//    to convert to array
//    e.g. {"a.0":"hello","a.1":"world"} will be converted to {"a":["hello","world"]}
//    note: the array index must be started from 0 and increment one by one
//          otherwise, it will not be converted to array but remain a key
//
// value rules:
//
// null value is ignored, therefore the associated path is also discarded
// bool value is converted to string
// string is directly mapped to go string
// number is wrapped as a Number (with string inside)
// array is converted to ConfigArray
// object is converted to ConfigObject
// a special object with single field '$ref' is converted to ConfigReference
//
type mapProcessor struct {
	root *ConfigObject
	m map[string]interface{}
}

func (mp *mapProcessor) setObject(path string, obj map[string]interface{}) error {
	// check if reference
	if r, exists := obj["$ref"]; exists && len(obj) == 1 {
		ref, ok := r.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid reference object %v", obj)
		}
		ls, exists := ref["labels"]
		if !exists {
			return fmt.Errorf("invalid reference object: labels not found")
		}
		labels, ok := ls.(Labels)
		if !ok {
			return fmt.Errorf("invalid reference object: labels is invalid %v", ls)
		}
		p, exists := ref["path"]
		if !exists {
			return fmt.Errorf("invalid reference object: path not found")
		}
		path, ok := p.(string)
		if !ok {
			return fmt.Errorf("invalid reference object: path is invalid %v", p)
		}
		cr := NewConfigReference(labels, path)
		mp.root.setReference(path, cr)
		return nil
	}
	co := NewConfigObject()
	mp.root.setObject(path, co)
	return mp.traverseObject(path, co, obj)
}

func (mp *mapProcessor) setArray(path string, arr []interface{}) error {
	ca := NewConfigArray()
	mp.root.setArray(path, ca)
	return mp.traverseArray(path, ca, arr)
}

func (mp *mapProcessor) setValue(path string, value interface{}) (err error) {
	switch value.(type) {
	case types.Nil: // ignored
	case string:
		mp.root.setString(path, value.(string))
	case bool:
		mp.root.setString(path, strconv.FormatBool(value.(bool)))
	case json.Number:
		mp.root.setNumber(path, FromJsonNumber(value.(json.Number)))
	case float64:
		mp.root.setNumber(path, Float64Number(value.(float64)))
	case []interface{}:
		err = mp.setArray(path, value.([]interface{}))
	case map[string]interface{}:
		err = mp.setObject(path, value.(map[string]interface{}))
	default:
		return fmt.Errorf("invalid value type: %T", value)
	}
	return
}

func (mp *mapProcessor) traverseArray(path string, ca *ConfigArray, arr []interface{}) (err error) {
	for idx, elem := range arr {
		switch elem.(type) {
		case types.Nil: // ignored
		case string:
			ca.addString(elem.(string))
		case bool:
			ca.addString(strconv.FormatBool(elem.(bool)))
		case json.Number:
			ca.addNumber(FromJsonNumber(elem.(json.Number)))
		case float64:
			ca.addNumber(Float64Number(elem.(float64)))
		case []interface{}:
			newpath := fmt.Sprintf("%v.%v", path, idx)
			newarr := elem.([]interface{})
			newca := NewConfigArray()
			ca.addArray(ca)
			if err = mp.traverseArray(newpath, newca, newarr); err != nil {
				return err
			}
		case map[string]interface{}:
			newpath := fmt.Sprintf("%v.%v", path, idx)
			newobj := elem.(map[string]interface{})
			// here we call setObject to handle reference internally
			if err = mp.setObject(newpath, newobj); err != nil {
				return err
			}
		}
	}
	return
}

func (mp *mapProcessor) traverseObject(path string, co *ConfigObject, obj map[string]interface{}) (err error) {
	if len(obj) == 0 {
		return nil
	}
	keys := make([]string, len(mp.m))
	i := 0
	for k := range mp.m {
		keys[i] = k
		i++
	}
	// order by key length
	sort.Sort(&stringLenSorter{keys})

	for _, k := range keys {
		value := mp.m[k]
		switch value.(type) {
		case types.Nil: // ignored
		case string:
			co.setString(k, value.(string))
		case bool:
			co.setString(k, strconv.FormatBool(value.(bool)))
		case json.Number:
			co.setNumber(k, FromJsonNumber(value.(json.Number)))
		case float64:
			co.setNumber(k, Float64Number(value.(float64)))
		case []interface{}:
			newpath := fmt.Sprintf("%v.%v", path, k)
			newca := NewConfigArray()
			newarr := value.([]interface{})
			co.setArray(k, newca)
			if err = mp.traverseArray(newpath, newca, newarr); err != nil {
				return err
			}
		case map[string]interface{}:
			newpath := fmt.Sprintf("%v.%v", path, k)
			if err = mp.setObject(newpath, value.(map[string]interface{})); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid value type: %T", value)
		}
	}
	return
}

func (mp *mapProcessor) traverse() error {
	if len(mp.m) == 0 {
		return nil
	}
	keys := make([]string, len(mp.m))
	i := 0
	for k := range mp.m {
		keys[i] = k
		i++
	}
	// order by key length
	sort.Sort(&stringLenSorter{keys})

	for _, k := range keys {
		v := mp.m[k]
		if err := mp.setValue(k, v); err != nil {
			return err
		}
	}
	return nil
}


func concatPath(parent string, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}