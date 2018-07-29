package kv

import (
	"testing"
	"encoding/json"
	"bytes"
	"fmt"
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
	err := json.Unmarshal(json1, &m1)
	if err != nil {
		t.Logf("expected error: %v", err)
	} else {
		t.Errorf("unmarshal error expected but not occur: %v", m1)
	}

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