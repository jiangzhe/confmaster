package kv

import (
	"testing"
	"encoding/json"
	"gopkg.in/yaml.v2"
)

var (
	kv1 = map[string]interface{}{
		"hello": "world",
		"ask": 123,
		"friends": []string{"john", "alice", "allen"},
	}
	//ref = &Reference{
	//	Namespace: "default",
	//	Labels: map[string]string{"project":"demo","cluster":"1","app":"nginx"},
	//	Path: "",
	//}
)

func TestJSON(t *testing.T) {
	bs, err := json.MarshalIndent(kv1, "", "  ")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("\n" + string(bs))
}

//func TestJSONReference(t *testing.T) {
//	bs, err := json.MarshalIndent(ref, "", "  ")
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	t.Logf("\n" + string(bs))
//}

func TestYAML(t *testing.T) {
	bs, err := yaml.Marshal(kv1)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("\n" + string(bs))
}

//func TestFormatImmutableRef(t *testing.T) {
//	NewJsonFormatter().Format()
//}

//func TestXML(t *testing.T) {
//	bs, err := xml.MarshalIndent(kv1, "", "  ")
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	t.Logf("\n" + string(bs))
//}