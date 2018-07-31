package kv

import (
	"k8s.io/apimachinery/pkg/util/intstr"
	"encoding/json"
	"errors"
	"bytes"
)

// definition is the metadata of a configuration for user
// to configure
type Definition struct {
	// path is a dot delimited string indicate the hierarchy of current config
	Path string `json:"path"`
	Type *ValueType `json:"type"`
	Description string `json:"description,omitempty"`
	Options []intstr.IntOrString `json:"options,omitempty"`
	Range *ValueRange `json:"range,omitempty"`
	Ref *ConfigReference `json:"ref,omitempty"`
}

var (
	ErrParseValueRange = errors.New("error parsing value range")
)

type ValueRange struct {
	Min Number `json:"min"`
	Max Number `json:"max"`
}

func (vr *ValueRange) MarshalJSON() ([]byte, error) {
	arr := []Number{vr.Min, vr.Max}
	return json.Marshal(arr)
}

func (vr *ValueRange) UnmarshalJSON(data []byte) error {
	if len(data) < 0 {
		return ErrParseValueRange
	}
	buf := bytes.NewBuffer(data)
	arr := make([]Number, 0)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err := decoder.Decode(&arr)
	if err != nil {
		return ErrParseValueRange
	}
	if len(arr) != 2 {
		return ErrParseValueRange
	}
	vr.Min = arr[0]
	vr.Max = arr[1]
	return nil
}

