package kv

import (
	"testing"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func TestConfigRefFromJson(t *testing.T) {
	json1 := []byte(`{"local":{"$ref":{"labels":{"cluster":"1"},"path":"remote"}}}`)
	conf, err := ConfigFromJson(json1)
	if err != nil {
		t.Errorf("json parse error: %v", err)
		return
	}
	ref := conf.GetReference("local")
	if ref == nil {
		t.Errorf("reference parse error")
		return
	}
	t.Logf("reference result: %T %v", ref, ref)
}

func TestConfigFromYaml(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("get current dir error: %v", err)
		return
	}

	file, err := os.Open(path.Clean(path.Join(dir, "testdata/application.yml")))
	if err != nil {
		t.Errorf("failed to open file %v", err)
		return
	}
	bs, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("failed to read file %v", err)
		return
	}
	conf, err := ConfigFromYaml(bs)
	if err != nil {
		t.Errorf("yaml parse error %v", err)
		return
	}

	t.Logf("keys: %v", len(conf.Keys()))
	t.Logf("refs: %v", len(conf.Refs()))
}

func TestConfigObject(t *testing.T) {
	co := NewConfigObject()
	co.setNumber("n", Int64Number(100))
	t.Logf(displayConfigObject(co))
	co.setString("a.b", "hello, world")
	t.Logf(displayConfigObject(co))
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

func TestMapFromJson(t *testing.T) {
	m, err := mapFromJson([]byte(`{"a":"A", "b":123, "c.d":{}, "e":null, "f":[1, true, null], "g":true}`))
	if err != nil {
		t.Errorf("json unmarshal error: %v", err)
	}
	t.Logf("map result: %v", m)
	t.Logf("type b: %T", m["b"])

	for k, v := range m {
		t.Logf("k=%v, v=%v, v.type=%T", k, v, v)
	}
}

func TestFallback1(t *testing.T) {
	conf, _ := ConfigFromJson([]byte(`{"a":"A","b":"B"}`))
	fb, _ := ConfigFromJson([]byte(`{"a":"AAA","c":"CCC"}`))
	mixed := conf.WithFallback(fb)
	// a should keep origin value
	astr := mixed.GetString("a")
	if astr != "A" {
		t.Errorf("a value mismatch: %v", astr)
		return
	}
	cstr := mixed.GetString("c")
	if cstr != "CCC" {
		t.Errorf("c value mismatch: %v", cstr)
		return
	}
}

// helper functions

func displayConfigObject(co *ConfigObject) string {
	buf := bytes.Buffer{}
	writeConfigObject(co, &buf)
	return buf.String()
}

func writeConfigObject(co *ConfigObject, buf *bytes.Buffer) string {
	buf.WriteByte('\n')
	refs := co.Refs()
	if len(refs) == 0 {
		buf.WriteString("refs: <nil>\n")
	} else {
		buf.WriteString(fmt.Sprintf("refs: %v\n", refs))
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

