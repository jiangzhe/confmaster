package kv

import (
	"testing"
	"encoding/json"
	"reflect"
)

func TestConfigRefFromJson(t *testing.T) {
	json1 := []byte(`
{
  "remote": {
    "$ref": {
      "namespace": "default",
      "labels": {
        "cluster": "1",
        "app": "nginx"
      },
      "path": "export.config"
    }
  },
  "remote.a": "A"
}
`)
	m1 := make(map[string]interface{})
	json.Unmarshal(json1, &m1)

	c := NewReadableConfig("", m1)

	_, err := c.Shrink()
	if err != nil {
		t.Logf("expected error: %v", err)
	} else {
		t.Error("path conflict not detected")
	}

	rc, _ := c.GetConfig("remote")
	t.Logf("getconfig(remote), %v", rc)
	if rc.Reference() == nil {
		t.Error("reference miss")
	}
}

func TestConfigRead(t *testing.T) {
	m := make(map[string]interface{})

	sm := make(map[string]interface{})
	sm["c"] = "C"
	sm["d"] = "D"

	m["a"] = "A"
	m["b"] = sm

	c := NewReadableConfig("app", m)
	t.Logf("path: %v", c.Path())
	if c.Path() != "app" {
		t.Error("path mismatch")
	}
	va, _ := c.Get("a")
	t.Logf("get(a): %v", va)
	if !reflect.DeepEqual(va, "A") {
		t.Error("get mismatch")
	}
	vb, _ := c.Get("b")
	t.Logf("get(b): %v", vb)
	if !reflect.DeepEqual(vb,  sm) {
		t.Error("get complex mismatch")
	}

	cb, _ := c.GetConfig("b")
	t.Logf("getconfig(b): %v", cb)
	vd, _ := cb.Get("d")
	t.Logf("get(d): %v", vd)
	if !reflect.DeepEqual(vd, "D") {
		t.Error("get value from sub config mismatch")
	}


}