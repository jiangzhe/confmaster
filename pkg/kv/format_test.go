package kv

import (
	"testing"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"errors"
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

type mockResolver struct {}

func (mr *mockResolver) Resolve(config ConfigInterface) (ResolvedConfigInterface, error) {
	switch config.(type) {
	case *ConfigObject:
		return &ResolvedConfigObject{m: config.(*ConfigObject).m}, nil
	}
	return nil, errors.New("unable to resolve config")
}