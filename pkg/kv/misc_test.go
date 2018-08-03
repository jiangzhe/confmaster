package kv

import (
	"testing"
	"reflect"
	"gopkg.in/yaml.v2"
	"encoding/json"
)

func TestNumberClone(t *testing.T) {
	i1 := Int64Number(100)
	i2 := i1.Clone()
	t.Logf("%p %p", i1, i2)
	*i1 = *Int64Number(200)
	t.Logf("%p %p", i1, i2)
	t.Logf("%v %v", i1, i2)
}

func TestDeepEqual(t *testing.T) {
	m1 := make(map[string]interface{})
	m1["a"] = 1
	m1["b"] = "B"
	var arr1 []interface{}
	for i := 1; i < 4; i++ {
		arr1 = append(arr1, i)
	}
	d1 := make(map[string]interface{})
	d1["d"] = arr1
	m1["c"] = d1
	m2 := make(map[string]interface{})
	m2["a"] = 1
	m2["b"] = "B"
	var arr2 []interface{}
	for i := 1; i < 4; i++ {
		arr2 = append(arr2, i)
	}
	d2 := make(map[string]interface{})
	d2["d"] = arr2
	m2["c"] = d2

	if !reflect.DeepEqual(m1, m2) {
		t.Errorf("not equal: %v %v", m1, m2)
	}
}


func TestJsonUnmarshal(t *testing.T) {
	m1 := make(map[string]interface{})
	json1 := []byte(`
{
  "a": {"b": 123},
  "a": {"c": 456}
}
`)
	json.Unmarshal(json1, &m1)
	t.Logf("result: %v", m1)
}

func TestYamlUnmarshal(t *testing.T) {
	ms := yaml.MapSlice{}
	yaml1 := []byte(`
a: hello
b:
  c: world
e:
  f: true

b:
  x: false
  10: 123

`)
	err := yaml.Unmarshal(yaml1, &ms)
	if err != nil {
		t.Errorf("unmarshal error: %v", err)
		return
	}
	arr := ms[3].Value

	t.Logf("%v", ms)
	t.Logf("%v", arr)
}

func TestRangeEmptyArr(t *testing.T) {
	arr := make(map[string]interface{})
	for k, v := range arr {
		t.Logf("key=%v, value=%v", k, v)
	}
}