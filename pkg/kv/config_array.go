package kv

type ConfigArray struct {
	arr []*Value
}

func (ca *ConfigArray) addString(value string) {
	ca.arr = append(ca.arr, MakeStringValue(value))
}

func (ca *ConfigArray) addNumber(value float64) {
	ca.arr = append(ca.arr, MakeNumericValue(value))
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