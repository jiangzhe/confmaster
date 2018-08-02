package kv

import (
	"errors"
	"encoding/json"
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"confmaster/pkg/kv/jsonpath"
	"strings"
	"strconv"
)

type ConfigInterface interface {
	GetValue(path string) *Value
	GetString(path string, defaultValue string) string
	GetBool(path string, defaultValue bool) bool
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

// read yaml bytes into ordered map slice and convert to config object
func ConfigFromYaml(input []byte) (*ConfigObject, error) {
	ms := new(yaml.MapSlice)
	err := yaml.Unmarshal(input, ms)
	if err != nil {
		return nil, err
	}
	return configFromMapSlice(ms)
}

func parseMapItemKey(itemKey interface{}) string {
	var key string
	switch k := itemKey.(type) {
	case string:
		key = k
	case int:
		key = strconv.Itoa(k)
	case json.Number:
		key = string(k)
	}
	return key
}

func parseMapItemValue(itemValue interface{}) (*Value, error) {
	var value *Value
	if itemValue == nil {
		return MakeStringValue(""), nil
	}
	switch v := itemValue.(type) {
	case int:
		value = MakeIntValue(int64(v))
	case float64:
		value = MakeFloatValue(v)
	case bool:
		value = MakeBoolValue(v)
	case string:
		value = MakeStringValue(v)
	case []interface{}:
		ca := NewConfigArray()
		for _, elem := range v {
			v, err := parseMapItemValue(elem)
			if err != nil {
				return nil, err
			}
			ca.arr = append(ca.arr, v)
		}
		value = MakeArrayValue(ca)
	case yaml.MapSlice:
		co := NewConfigObject()
		for _, item := range v {
			subKey := parseMapItemKey(item.Key)
			if subKey == "" {
				return nil, fmt.Errorf("invalid key to parse: %v", item.Key)
			}
			subValue, err := parseMapItemValue(item.Value)
			if err != nil {
				return nil, err
			}
			if err := co.setValue(subKey, subValue); err != nil {
				return nil, err
			}
		}
		value = MakeObjectValue(co)
	}
	if value == nil {
		return nil, fmt.Errorf("invalid valur to parse: %T %v", itemValue, itemValue)
	}
	return value, nil
}

func configFromMapSlice(ms *yaml.MapSlice) (*ConfigObject, error) {
	co := NewConfigObject()
	if len(*ms) == 0 {
		return co, nil
	}
	for _, item := range *ms {
		subKey := parseMapItemKey(item.Key)
		if subKey == "" {
			return nil, fmt.Errorf("invalid key to parse: %v", item.Key)
		}
		subValue, err := parseMapItemValue(item.Value)
		if err != nil {
			return nil, err
		}
		if err := co.setValue(subKey, subValue); err != nil {
			return nil, err
		}
	}
	return co, nil
}

func ConfigFromJson(input []byte) (*ConfigObject, error) {
	buf := bytes.NewBuffer(input)
	return ConfigFromJsonStream(buf)
}

// reuse MapSlice to deal with duplicate keys in both json and yaml
func ConfigFromJsonStream(reader io.Reader) (*ConfigObject, error) {
	decoder := jsonpath.NewDecoder(reader)
	decoder.UseNumber()

	co := NewConfigObject()

	err := decoder.Scan(func(path jsonpath.JsonPath, value json.Token) error {
		if value == nil {
			return co.setValue(path.String(), MakeStringValue(""))
		}
		switch value := value.(type) {
		case float64: return co.setValue(path.String(), MakeFloatValue(value))
		case bool: return co.setValue(path.String(), MakeBoolValue(value))
		case json.Number: return co.setValue(path.String(), makeNumericValue(FromJsonNumber(value)))
		case int: return co.setValue(path.String(), MakeIntValue(int64(value)))
		case string: return co.setValue(path.String(), MakeStringValue(value))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// copy the config object and convert $ref to reference
	cloned := co.Clone()
	err = co.traverse(func(path string, value *Value) error {
		switch value.Type {
		case ObjectType:
			if strings.HasSuffix(path, ".$ref") {
				cr, err := parseReference(value.RefValue.(*ConfigObject))
				if err != nil {
					return err
				}
				refpath := path[:len(path)-5]
				if err := cloned.setReference(refpath, cr); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return cloned, nil
}

func parseReference(ref *ConfigObject) (*ConfigReference, error) {
	ls := ref.GetObject("labels")
	if ls == nil {
		return nil, fmt.Errorf("invalid reference object: missing labels")
	}
	keys := ls.Keys()
	labels := make(map[string]string)
	for _, key := range keys {
		value := ls.GetValue(key)
		if value == nil || value.Type != StringType {
			return nil, fmt.Errorf("invalid reference object: invalid value of key '%T %v' in labels", key, key)
		}
		labels[key] = value.RefValue.(string)
	}

	p := ref.GetValue("path")
	if p == nil || p.Type != StringType {
		return nil, fmt.Errorf("invalid reference object: missing or invalid path '%T %v'", p, p)
	}
	path := p.RefValue.(string)
	return NewConfigReference(Labels(labels), path), nil
}


func mapFromYaml(input []byte) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := yaml.Unmarshal(input, &m)
	return m, err
}

//
//// parse config object from map,
//// if map is empty, return nil
////
//// key rules:
////
//// 1. key must be string
//// 2. if key has dot inside, it is treated as a path composite of several keys
////    path will be split and the value will be set through it, leading to
////    a nested structure generated
////    e.g. {"a.b.c":"hello"} will be converted to {"a":{"b":{"c":"hello"}}}
//// 3. key pattern like 'key[n]' indicates one element in array, the number in brackets
////    is the index of element
////    e.g. {"a[0]":"hello","a[1]":"world"} will be converted to {"a":["hello","world"]}
////    note: the array index must be started from 0 and increment one by one
////          otherwise, error will be thrown
////
//// value rules:
////
//// null value is ignored, therefore the associated path is also discarded
//// bool value is converted to string
//// string is directly mapped to go string
//// number is wrapped as a Number (with string inside)
//// array is converted to ConfigArray
//// object is converted to ConfigObject
//// a special object with single field '$ref' is converted to ConfigReference
////
//type mapProcessor struct {
//	root *ConfigObject
//	m map[string]interface{}
//}
//
//func (mp *mapProcessor) traverse() error {
//	if len(mp.m) == 0 {
//		return nil
//	}
//	keys := make([]string, len(mp.m))
//	i := 0
//	for k := range mp.m {
//		keys[i] = k
//		i++
//	}
//	// order by key length
//	sort.Sort(&stringLenSorter{keys})
//
//	for _, k := range keys {
//		v := mp.m[k]
//		if err := mp.setValue(k, v); err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
////
////func parseReference(ref interface{}) (*Value, error) {
////	refmap, ok := ref.(map[string]interface{})
////	if !ok {
////		return nil, fmt.Errorf("invalid reference object: %v", ref)
////	}
////	ls, exists := refmap["labels"]
////	if !exists {
////		return nil, fmt.Errorf("invalid reference object: missing labels")
////	}
////	tmplabels, ok := ls.(map[string]interface{})
////	if !ok {
////		return nil, fmt.Errorf("invalid reference object: invalid labels %v", ls)
////	}
////	labels := make(map[string]string, len(tmplabels))
////	for k, v := range tmplabels {
////		if s, ok := v.(string); !ok {
////			return nil, fmt.Errorf("invalid reference object, invalid value in labels %T %v", v, v)
////		} else {
////			labels[k] = s
////		}
////	}
////
////	p, exists := refmap["path"]
////	if !exists {
////		return nil, fmt.Errorf("invalid reference object: missing path")
////	}
////	path, ok := p.(string)
////	if !ok {
////		return nil, fmt.Errorf("invalid reference object: invalid path %v", p)
////	}
////	return MakeReferenceValue(NewConfigReference(Labels(labels), path)), nil
////}
//
//// need to adapt results of json/yaml parsing
//func wrapValue(value interface{}) (*Value, error) {
//	if value == nil {
//		return MakeStringValue(""), nil
//	}
//	switch value.(type) {
//	case types.Nil: return MakeStringValue(""), nil
//	case string: return MakeStringValue(value.(string)), nil
//	case bool: return MakeBoolValue(value.(bool)), nil
//	case json.Number: return makeNumericValue(FromJsonNumber(value.(json.Number))), nil
//	case float64: return makeNumericValue(Float64Number(value.(float64))), nil
//	case int: return makeNumericValue(Int64Number(int64(value.(int)))), nil
//	case []interface{}:
//		arr := value.([]interface{})
//		ca := NewConfigArray()
//		for idx, elem := range arr {
//			v, err := wrapValue(elem)
//			if err != nil {
//				return nil, err
//			}
//			ca.setValue(idx, v)
//		}
//		return MakeArrayValue(ca), nil
//	case map[string]interface{}:
//		m := value.(map[string]interface{})
//		if ref, exists := m["$ref"]; exists {
//			return parseReference(ref)
//		}
//		co := NewConfigObject()
//		for k, v := range m {
//			wrapped, err := wrapValue(v)
//			if err != nil {
//				return nil, err
//			}
//			co.setValue(k, wrapped)
//		}
//		return MakeObjectValue(co), nil
//	// do not use
//	//case map[interface{}]interface{}:
//	//	m := value.(map[interface{}]interface{})
//	//	if ref, exists := m["$ref"]; exists {
//	//		return parseReference(ref)
//	//	}
//	//	co := NewConfigObject()
//	//	for key, v := range m {
//	//		k, ok := key.(string)
//	//		if !ok {
//	//			return nil, fmt.Errorf("invalid key type: %T", key)
//	//		}
//	//		wrapped, err := wrapValue(v)
//	//		if err != nil {
//	//			return nil, err
//	//		}
//	//		co.setValue(k, wrapped)
//	//	}
//	//	return MakeObjectValue(co), nil
//	default:
//		return nil, fmt.Errorf("invalid value type: %T", value)
//	}
//}
//
//func (mp *mapProcessor) setValue(path string, value interface{}) error {
//	v, err := wrapValue(value)
//	if err != nil {
//		return err
//	}
//	return mp.root.setValue(path, v)
//}
