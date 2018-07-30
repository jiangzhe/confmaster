package kv

import (
	"testing"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"errors"
	"reflect"
)

var (
	kv1 = map[string]interface{}{
		"hello": "world",
		"ask": 123,
		"friends": []string{"john", "alice", "allen"},
	}
)

func TestJSON(t *testing.T) {
	bs, err := json.MarshalIndent(kv1, "", "  ")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("\n" + string(bs))
}

func TestYAML(t *testing.T) {
	bs, err := yaml.Marshal(kv1)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("\n" + string(bs))
}

func mockConfigObject() ResolvedConfigInterface {
	co := NewConfigObject()
	co.setInt("a", 1)
	co.setString("b", "B")

	ca := NewConfigArray()
	ca.addInt(1)
	ca.addInt(2)
	ca.addInt(3)

	co.setArray("c.d", ca)
	mr := &mockResolver{}
	rco, _ := mr.Resolve(co)
	return rco
}

func TestConfigObjectFormatJson(t *testing.T) {
	rco := mockConfigObject()
	bs, err := rco.Format(NewJsonFormatter(0, 2))
	if err != nil {
		t.Errorf("format error: %v", err)
		return
	}
	t.Logf("json:\n%v", string(bs))
	m1 := make(map[string]interface{})
	json.Unmarshal(bs, &m1)
	m2 := rco.ToMap()
	if !reflect.DeepEqual(m1, m2) {
		t.Errorf("format mismatch: %v %v", m1, m2)
		d1 := m1["c"].(map[string]interface{})["d"]
		d2 := m2["c"].(map[string]interface{})["d"]
		t.Logf("type d: %t %t", d1, d2)
		n2, err := d2.(float64)
		t.Logf("n2=%v, err=%v", n2, err)
	}
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

func TestConfigObjectFormatYaml(t *testing.T) {

}

type mockResolver struct {}

func (mr *mockResolver) Resolve(config ConfigInterface) (ResolvedConfigInterface, error) {
	switch config.(type) {
	case *ConfigObject:
		return &ResolvedConfigObject{m: config.(*ConfigObject).m}, nil
	}
	return nil, errors.New("unable to resolve config")
}