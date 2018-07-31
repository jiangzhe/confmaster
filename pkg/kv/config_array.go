package kv

import "strconv"

type ConfigArray struct {
	arr []*Value
}

func (ca *ConfigArray) addString(value string) {
	ca.arr = append(ca.arr, MakeStringValue(value))
}

func (ca *ConfigArray) addNumber(value *Number) {
	ca.arr = append(ca.arr, makeNumericValue(value))
}

func (ca *ConfigArray) addInt(value int64) {
	ca.arr = append(ca.arr, MakeIntValue(value))
}

func (ca *ConfigArray) addFloat(value float64) {
	ca.arr = append(ca.arr, MakeFloatValue(value))
}

func (ca *ConfigArray) addObject(value *ConfigObject) {
	ca.arr = append(ca.arr, MakeObjectValue(value))
}

func (ca *ConfigArray) addArray(value *ConfigArray) {
	ca.arr = append(ca.arr, MakeArrayValue(value))
}

func (ca *ConfigArray) addReference(value *ConfigReference) {
	ca.arr = append(ca.arr, MakeReferenceValue(value))
}

func (ca *ConfigArray) addValue(t ValueType, value interface{}) {
	ca.arr = append(ca.arr, &Value{Type: t, RefValue: value})
}

func (ca *ConfigArray) setString(path string, value string) error {
	return setValue(ca, path, MakeStringValue(value))
}

func (ca *ConfigArray) setNumber(path string, value *Number) error {
	return setValue(ca, path, makeNumericValue(value))
}

// todo merge object
func (ca *ConfigArray) setObject(path string, value *ConfigObject) error {
	return setValue(ca, path, MakeObjectValue(value))
}

// todo merge array ? maybe not
func (ca *ConfigArray) setArray(path string, value *ConfigArray) error {
	return setValue(ca, path, MakeArrayValue(value))
}

func (ca *ConfigArray) setReference(path string, value *ConfigReference) error {
	return setValue(ca, path, MakeReferenceValue(value))
}

func (ca *ConfigArray) getValueByKey(key string) *Value {
	n, err := strconv.Atoi(key)
	if err != nil {
		return nil
	}
	if n >= 0 && n < len(ca.arr) {
		return ca.arr[n]
	}
	return nil
}

func (ca *ConfigArray) unset(path string) error {
	// todo
}


func NewConfigArray() *ConfigArray {
	co := ConfigArray{
		arr: make([]*Value, 0),
	}
	return &co
}

func (ca *ConfigArray) Clone() *ConfigArray {
	arr := make([]*Value, len(ca.arr))
	for i, v := range ca.arr {
		arr[i] = v.Clone()
	}
	return &ConfigArray{
		arr: arr,
	}
}