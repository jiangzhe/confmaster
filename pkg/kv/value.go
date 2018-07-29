package kv

type ValueType int

const (
	StringType         ValueType = iota
	ReferenceType
	NumericType
	ObjectType
	ArrayType
	FallbackType
)

type Value struct {
	Type     ValueType
	RefValue interface{}
}

func (v *Value) cloneFrom(that *Value) {
	v.RefValue = that.RefValue
	v.Type = that.Type
}

func (v *Value) Clone() *Value {
	var ref interface{}
	switch v.Type {
	case ObjectType:
		ref = v.RefValue.(*ConfigObject).Clone()
	case ArrayType:
		ref = v.RefValue.(*ConfigArray).Clone()
	case ReferenceType:
		ref = v.RefValue.(*ConfigReference).Clone()
	case StringType:
	case NumericType:
		ref = v.RefValue
	}
	return &Value{
		Type: v.Type,
		RefValue: ref,
	}
}

type valueSetter interface {
	setString(name string, value string)
	setInt(name string, value int)
	setObject(name string, value *ConfigObject)
	setArray(name string, value *ConfigArray)
	setReference(name string, value *ConfigReference)
	setValue(path string, t ValueType, ref interface{})
}

func MakeStringValue(src string) *Value {
	return &Value{
		Type:     StringType,
		RefValue: src,
	}
}

func MakeNumericValue(src int) *Value {
	return &Value{
		Type:     NumericType,
		RefValue: src,
	}
}

func MakeObjectValue(src *ConfigObject) *Value {
	return &Value{
		Type:     ObjectType,
		RefValue: src,
	}
}

func MakeArrayValue(src *ConfigArray) *Value {
	return &Value{
		Type:     ArrayType,
		RefValue: src,
	}
}

func MakeReferenceValue(src *ConfigReference) *Value {
	return &Value{
		Type:     ReferenceType,
		RefValue: src,
	}
}

func MakeFallbackValue(src *ConfigFallback) *Value {
	return &Value{
		Type:     FallbackType,
		RefValue: src,
	}
}
