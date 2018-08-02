package kv

import (
	"testing"
	"errors"
	"io/ioutil"
	"os"
	"path"
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
}

func TestConfigObjectFormatYaml(t *testing.T) {
	rco := mockConfigObject()
	bs, err := rco.Format(NewYamlFormatter(0))
	if err != nil {
		t.Errorf("format error: %v", err)
		return
	}
	t.Logf("yaml:\n%v", string(bs))
}

func TestConfigObjectReformatYaml(t *testing.T) {
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
	mr := &mockResolver{}
	rco, err := mr.Resolve(conf)
	if err != nil {
		t.Errorf("resolve error %v", err)
		return
	}
	bs, err = rco.Format(NewYamlFormatter(0))
	if err != nil {
		t.Errorf("format error: %v", err)
		return
	}
	t.Logf("yaml:\n%v", string(bs))
}

type mockResolver struct {}

func (mr *mockResolver) Resolve(config ConfigInterface) (ResolvedConfigInterface, error) {
	switch config.(type) {
	case *ConfigObject:
		return &ResolvedConfigObject{m: config.(*ConfigObject).m}, nil
	}
	return nil, errors.New("unable to resolve config")
}