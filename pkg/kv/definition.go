package kv

import (
	"k8s.io/apimachinery/pkg/util/intstr"
	"encoding/json"
	"errors"
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
	min float64
	max float64
}

func (vr *ValueRange) MarshalJSON() ([]byte, error) {
	arr := []float64{vr.min, vr.max}
	return json.Marshal(arr)
}

func (vr *ValueRange) UnmarshalJSON(data []byte) error {
	if len(data) < 0 {
		return ErrParseValueRange
	}
	arr := make([]float64, 0)
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return ErrParseValueRange
	}
	if len(arr) != 2 {
		return ErrParseValueRange
	}
	vr.min = arr[0]
	vr.max = arr[1]
	return nil
}

