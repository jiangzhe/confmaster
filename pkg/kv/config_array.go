package kv

import "fmt"

type ConfigArray struct {
	arr []*Value
}

func (ca *ConfigArray) getValue(idx int) *Value {
	if idx < 0 || idx >= len(ca.arr) {
		return nil
	}
	return ca.arr[idx]
}

// set value to specified index
// should take care of object merge
func (ca *ConfigArray) setValue(idx int, value *Value) error {
	if idx < 0 || idx > len(ca.arr) {
		return fmt.Errorf("index out of bound: %v", idx)
	}
	if idx == len(ca.arr) {
		ca.arr = append(ca.arr, value)
		return nil
	}
	// consider object merge
	origvalue := ca.arr[idx]
	if origvalue.Type == ObjectType && value.Type == ObjectType {
		origco := origvalue.RefValue.(*ConfigObject)
		newco := value.RefValue.(*ConfigObject)
		return origco.merge(newco)
	}

	ca.arr[idx] = value
	return nil
}

func (ca *ConfigArray) setString(idx int, value string) error {
	return ca.setValue(idx, MakeStringValue(value))
}

func (ca *ConfigArray) setNumber(idx int, value *Number) error {
	return ca.setValue(idx, makeNumericValue(value))
}

func (ca *ConfigArray) setInt(idx int, value int64) error {
	return ca.setValue(idx, MakeIntValue(value))
}

func (ca *ConfigArray) setFloat(idx int, value float64) error {
	return ca.setValue(idx, MakeFloatValue(value))
}

func (ca *ConfigArray) setObject(idx int, value *ConfigObject) error {
	return ca.setValue(idx, MakeObjectValue(value))
}

func (ca *ConfigArray) setArray(idx int, value *ConfigArray) error {
	return ca.setValue(idx, MakeArrayValue(value))
}

func (ca *ConfigArray) setReference(idx int, value *ConfigReference) error {
	return ca.setValue(idx, MakeReferenceValue(value))
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