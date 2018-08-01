package kv

import (
	"testing"
	"encoding/json"
	"errors"
	"reflect"
)

func mockConfigObject() ResolvedConfigInterface {
	co := NewConfigObject()
	co.setNumber("a", Int64Number(1))
	co.setString("b", "B")

	ca := NewConfigArray()
	ca.setInt(0, 1)
	ca.setInt(1, 2)
	ca.setInt(2, 3)

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
	//buf := bytes.NewBuffer(bs)
	//decoder := json.NewDecoder(buf)
	//decoder.UseNumber()
	//decoder.Decode(&m1)
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

type mockResolver struct {}

func (mr *mockResolver) Resolve(config ConfigInterface) (ResolvedConfigInterface, error) {
	switch config.(type) {
	case *ConfigObject:
		return &ResolvedConfigObject{m: config.(*ConfigObject).m}, nil
	}
	return nil, errors.New("unable to resolve config")
}