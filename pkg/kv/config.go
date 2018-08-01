package kv

import (
	"errors"
	"encoding/json"
	"bytes"
	"go/types"
	"strconv"
	"sort"
	"fmt"
	"gopkg.in/yaml.v2"
)

type ConfigInterface interface {
	GetValue(path string) *Value
	GetString(path string) string
	GetNumber(path string) *Number
	GetObject(path string) *ConfigObject
	GetArray(path string) *ConfigArray
	GetReference(path string) *ConfigReference
	Refs() map[string]*ConfigReference
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
func ConfigFromJson(input []byte) (*ConfigObject, error) {
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

// read yaml bytes and returns config
func ConfigFromYaml(input []byte) (*ConfigObject, error) {
	m, err := mapFromYaml(input)
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

func mapFromJson(input []byte) (map[string]interface{}, error) {
	buf := bytes.NewBuffer(input)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	m := make(map[string]interface{})
	err := decoder.Decode(&m)
	return m, err
}

func mapFromYaml(input []byte) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := yaml.Unmarshal(input, &m)
	return m, err
}

// parse config object from map,
// if map is empty, return nil
//
// key rules:
//
// 1. key must be string
// 2. if key has dot inside, it is treated as a path composite of several keys
//    path will be split and the value will be set through it, leading to
//    a nested structure generated
//    e.g. {"a.b.c":"hello"} will be converted to {"a":{"b":{"c":"hello"}}}
// 3. key pattern like 'key[n]' indicates one element in array, the number in brackets
//    is the index of element
//    e.g. {"a[0]":"hello","a[1]":"world"} will be converted to {"a":["hello","world"]}
//    note: the array index must be started from 0 and increment one by one
//          otherwise, error will be thrown
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


func parseReference(ref interface{}) (*Value, error) {
	refmap, ok := ref.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid reference object: %v", ref)
	}
	ls, exists := refmap["labels"]
	if !exists {
		return nil, fmt.Errorf("invalid reference object: missing labels")
	}
	tmplabels, ok := ls.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid reference object: invalid labels %v", ls)
	}
	labels := make(map[string]string, len(tmplabels))
	for k, v := range tmplabels {
		if s, ok := v.(string); !ok {
			return nil, fmt.Errorf("invalid reference object, invalid value in labels %T %v", v, v)
		} else {
			labels[k] = s
		}
	}

	p, exists := refmap["path"]
	if !exists {
		return nil, fmt.Errorf("invalid reference object: missing path")
	}
	path, ok := p.(string)
	if !ok {
		return nil, fmt.Errorf("invalid reference object: invalid path %v", p)
	}
	return MakeReferenceValue(NewConfigReference(Labels(labels), path)), nil
}

// need to adapt results of json/yaml parsing
func wrapValue(value interface{}) (*Value, error) {
	if value == nil {
		return MakeStringValue(""), nil
	}
	switch value.(type) {
	case types.Nil: return MakeStringValue(""), nil
	case string: return MakeStringValue(value.(string)), nil
	case bool: return MakeStringValue(strconv.FormatBool(value.(bool))), nil
	case json.Number: return makeNumericValue(FromJsonNumber(value.(json.Number))), nil
	case float64: return makeNumericValue(Float64Number(value.(float64))), nil
	case int: return makeNumericValue(Int64Number(int64(value.(int)))), nil
	case []interface{}:
		arr := value.([]interface{})
		ca := NewConfigArray()
		for idx, elem := range arr {
			v, err := wrapValue(elem)
			if err != nil {
				return nil, err
			}
			ca.setValue(idx, v)
		}
		return MakeArrayValue(ca), nil
	case map[string]interface{}:
		m := value.(map[string]interface{})
		if ref, exists := m["$ref"]; exists {
			return parseReference(ref)
		}
		co := NewConfigObject()
		for k, v := range m {
			wrapped, err := wrapValue(v)
			if err != nil {
				return nil, err
			}
			co.setValue(k, wrapped)
		}
		return MakeObjectValue(co), nil
	case map[interface{}]interface{}:
		m := value.(map[interface{}]interface{})
		if ref, exists := m["$ref"]; exists {
			return parseReference(ref)
		}
		co := NewConfigObject()
		for key, v := range m {
			k, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("invalid key type: %T", key)
			}
			wrapped, err := wrapValue(v)
			if err != nil {
				return nil, err
			}
			co.setValue(k, wrapped)
		}
		return MakeObjectValue(co), nil
	default:
		return nil, fmt.Errorf("invalid value type: %T", value)
	}
}

func (mp *mapProcessor) setValue(path string, value interface{}) error {
	v, err := wrapValue(value)
	if err != nil {
		return err
	}
	return mp.root.setValue(path, v)
}
