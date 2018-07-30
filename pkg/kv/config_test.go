package kv

import (
	"testing"
	"encoding/json"
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
)

func TestConfigRefFromJson(t *testing.T) {

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
	m1 := make(map[string]interface{})

	yaml1 := []byte(`
a:
  b: 123

a:
  c: 456

a.d: 789
`)
	yaml.Unmarshal(yaml1, &m1)
	t.Logf("result: %v", m1)
}

func TestDotJson(t *testing.T) {
	json1 := []byte(`{"hello.world":"perfect"}`)
	m1 := make(map[string]interface{})
	json.Unmarshal(json1, &m1)
	t.Logf("result: %v", m1)
}


func TestConfigObject(t *testing.T) {
	co := NewConfigObject()
	co.setInt("n", 100)
	t.Logf(displayConfigObject(co))
	co.setString("a.b", "hello, world")
	t.Logf(displayConfigObject(co))
	co.setReference("a", NewConfigReference(map[string]string{"cluster":"1", "app":"nginx"}, "consumer.config"))
	t.Logf(displayConfigObject(co))
}

func displayConfigObject(co *ConfigObject) string {
	buf := bytes.Buffer{}
	writeConfigObject(co, &buf)
	return buf.String()
}

func writeConfigObject(co *ConfigObject, buf *bytes.Buffer) string {
	buf.WriteByte('\n')
	if len(co.refs) == 0 {
		buf.WriteString("refs: <nil>\n")
	} else {
		buf.WriteString(fmt.Sprintf("refs: %v\n", co.refs))
	}

	if len(*co.m) == 0 {
		buf.WriteString("m: <nil>\n")
	} else {
		buf.WriteString("m: \n")
		for k, v := range *co.m {
			buf.WriteString(fmt.Sprintf("%v -> {type=%v value=%v}\n", k, v.Type, v.RefValue))
		}
	}
	return buf.String()
}

func TestConfigOverwritePath(t *testing.T) {
	co := NewConfigObject()
	co.setString("a", "A")
	co.setString("a.b", "B")
	vb := co.GetValue("a.b")
	if vb == nil {
		t.Error("vb is not expected as nil")
		return
	}
	if vb.Type != StringType {
		t.Errorf("vb has invalid type %v", vb.Type)
		return
	}

	va := co.GetObject("a")
	if va == nil {
		t.Error("va is not expected as nil")
		return
	}
}

