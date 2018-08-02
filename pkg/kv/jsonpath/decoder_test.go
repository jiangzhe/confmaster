package jsonpath

import (
	"testing"
	"bytes"
	"encoding/json"
)

func TestScan(t *testing.T) {
	json1 := []byte(`{"a":"A","b":{"c":true,"d":[1,2,3},"x":null}`)
	buf := bytes.NewReader(json1)
	decoder := NewDecoder(buf)
	decoder.Scan(func(path JsonPath, token json.Token) error {
		t.Logf("path: %v, token: %T %v", path, token, token)
		return nil
	})
}

func TestPathString(t *testing.T) {
	path := JsonPath([]interface{}{"hello", 1, "world"})
	t.Logf("path: %v", path.String())
}