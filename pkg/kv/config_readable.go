package kv

import (
	"errors"
	"sort"
	"strings"
	"fmt"
	"regexp"
	"github.com/mitchellh/mapstructure"
)

// config only for read
type ConfigReadable interface {
	// the path tells how to get current config from root
	// the root config has 0-length array as its root
	Path() string

	// get value by key
	Get(key string) (interface{}, bool)

	// get all keys
	Keys() []string

	//// get all values
	//Values() []interface{}

	// get and cast the value as config, useful for nested/structured key value settings
	GetConfig(key string) (ConfigReadable, error)

	// whether the config is a reference to other app's config
	// if it's a reference, it should belong to some mapping of one dependency
	// note modification is disabled(return error) if it is reference
	Reference() *Reference

	// return a raw map of its key values, the modification on the map may
	// or may not impact the config, depending on the implementation
	// users are recommended to use the resolved config for reading
	// and the writable config for writing
	RawMap() map[string]interface{}

	// shrink will traverse all key values in the config and merge values with same key
	// all the keys contains dot will be split and reinserted to nested structure,
	// except those that points to a reference
	Shrink() (ConfigReadable, error)

	// flatten all key values to be in top level, the key string is delimited by dot
	Flatten() ConfigReadable
}

// map based struct only implements ConfigReadable
type readableConfig struct {
	path string
	m map[string]interface{}
	ref *Reference
}

func NewReadableConfig(path string, m map[string]interface{}) ConfigReadable {
	return &readableConfig{
		path: path,
		m: m,
	}
}

func NewReferenceConfig(path string, ref *Reference) ConfigReadable {
	return &readableConfig{
		path: path,
		ref: ref,
	}
}

func (rc *readableConfig) Path() string {
	return rc.path
}

func (rc *readableConfig) Get(key string) (value interface{}, exists bool) {
	value, exists = rc.m[key]
	return
}

func (rc *readableConfig) Keys() []string {
	keys := make([]string, 0, len(rc.m))
	for k := range rc.m {
		keys = append(keys, k)
	}
	return keys
}

func (rc *readableConfig) GetConfig(key string) (ConfigReadable, error) {
	value, exists := rc.m[key]
	if !exists {
		return nil, ErrConfigNotExists
	}
	if m, ok := value.(map[string]interface{}); ok {
		var path string
		if len(rc.path) == 0 {
			path = key
		} else {
			path = rc.path + "." + key
		}

		var ref *Reference
		value, ok := m["$ref"]
		if !ok {
			return &readableConfig{ path: path, m: m }, nil
		} else {
			err := mapstructure.Decode(value, &ref)
			if err != nil {
				return &readableConfig{ path: path, m: m }, nil
			}
			return &readableConfig{ path: path, ref: ref }, nil
		}
	}

	return nil, ErrConfigCastInvalid
}

func (rc *readableConfig) Reference() *Reference {
	return rc.ref
}

func (rc *readableConfig) RawMap() map[string]interface{} {
	return rc.m
}

var (
	ArraySubscript = regexp.MustCompile("^.+\\[0-9+\\]$")
)


func (rc *readableConfig) Shrink() (ConfigReadable, error) {
	if rc.ref != nil {
		return NewReferenceConfig(rc.path, rc.ref), nil
	}

	m := make(map[string]interface{})
	keys := rc.Keys()
	// sort keys by lengths
	sort.Sort(&stringLenSorter{arr: keys})

	for _, k := range keys {
		v, _ := rc.m[k]
		err := addNestPathToMap(m, k, v)
		if err != nil {
			return nil, err
		}
	}

	return NewReadableConfig(rc.path, m), nil
}

func (rc *readableConfig) Flatten() ConfigReadable {
	if rc.ref != nil {
		return NewReferenceConfig(rc.path, rc.ref)
	}
	m := make(map[string]interface{})
	keys := rc.Keys()
	sort.Sort(&stringLenSorter{arr: keys})

	for _, k := range keys {
		v, _ := rc.m[k]
		addFlatPathToMap(m, k, v)
	}

	return &readableConfig{path: rc.path, m: m, ref: rc.ref}
}

type stringLenSorter struct {
	arr []string
}

func (sls *stringLenSorter) Len() int {
	return len(sls.arr)
}

func (sls *stringLenSorter) Less(i, j int) bool {
	return len(sls.arr[i]) < len(sls.arr[j])
}

func (sls *stringLenSorter) Swap(i, j int) {
	sls.arr[i], sls.arr[j] = sls.arr[j], sls.arr[i]
}


var (
	ErrConfigNotExists = errors.New("config not exists")
	ErrConfigChangeNotAllowed = errors.New("config change not allowed")
	ErrConfigCastInvalid = errors.New("config cast invalid")
	ErrValueReferenceConflict = errors.New("value reference conflict")
	ErrValuePathConflict = errors.New("value path conflict")
)

// helper functions

// add given p, v into map
// p is a dot delimited path
// structure is flatten
func addFlatPathToMap(m map[string]interface{}, path string, value interface{}) {
	if strings.HasSuffix(path, "$ref") {
		m[path] = value
		return
	}

	if mm, ok := value.(map[string]interface{}); ok {
		for k, v := range mm {
			addFlatPathToMap(m, path + "." + k, v)
		}
		return
	}

	if arr, ok := value.([]interface{}); ok {
		for i, v := range arr {
			addFlatPathToMap(m, fmt.Sprintf("%v[%v]", path, i), v)
		}
		return
	}

	m[path] = value
}

// add given p, v into map
// p is a dot delimited path
// structure is nested
func addNestPathToMap(m map[string]interface{}, p string, v interface{}) error {
	if strings.HasSuffix(p, "$ref") {
		m[p] = v
		return nil
	}

	dotIdx := strings.Index(p, ".")
	if dotIdx == -1 {
		m[p] = v
		return nil
	}

	prefix := p[:dotIdx]
	suffix := p[dotIdx+1:]
	ov, exists := m[prefix]
	if exists {
		sm, ok := ov.(map[string]interface{})
		if !ok {
			//sm = make(map[string]interface{})
			//sm[prefix] = sm
			//addNestPathToMap(sm, suffix, v)
			//return nil
			return ErrValuePathConflict
		}

		// if sm is reference, error out
		if _, exists := sm["$ref"]; exists {
			return ErrValueReferenceConflict
		}
		addNestPathToMap(sm, suffix, v)
		return nil
	}

	sm := make(map[string]interface{})
	addNestPathToMap(sm, suffix, v)
	m[prefix] = sm
	return nil
}
