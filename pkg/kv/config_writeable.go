package kv

// writable config provides methods to overwrite value
// on any path
//
type ConfigWritable interface {
	ConfigReadable

	Overwrite(path string, value interface{}) error

	OverwriteKeys() []string

	OverwriteConfig(path string, config ConfigReadable) error

	Merge() (ConfigWritable, error)
}

type writableConfig struct {
	origin readableConfig
	overwrite readableConfig
}

func (wc *writableConfig) Path() string {
	return wc.origin.path
}

func (wc *writableConfig) Get(key string) (value interface{}, exists bool) {
	value, exists = wc.overwrite.m[key]
	if exists {
		return
	}
	value, exists = wc.origin.m[key]
	return
}

func (wc *writableConfig) Keys() []string {
	mm := make(map[string]bool, len(wc.origin.m) + len(wc.overwrite.m))
	for k := range wc.origin.m {
		mm[k] = true
	}
	for k := range wc.overwrite.m {
		mm[k] = true
	}
	keys := make([]string, 0, len(mm))
	for k := range mm {
		keys = append(keys, k)
	}
	return keys
}

func (wc *writableConfig) GetConfig(key string) (ConfigReadable, error) {
	value, exists := wc.overwrite.m[key]
	if !exists {
		return wc.origin.GetConfig(key)
	}
	if m, ok := value.(map[string]interface{}); ok {
		var path string
		if len(wc.overwrite.path) == 0 {
			path = key
		} else {
			path = wc.overwrite.path + "." + key
		}

		var ref *Reference
		if value, ok := m["$ref"]; !ok {
			return &readableConfig{ path: path, m: m }, nil
		} else if ref, ok = value.(*Reference); !ok {
			return &readableConfig{ path: path, m: m }, nil
		}
		return &readableConfig{ path: path, ref: ref }, nil
	}

	return nil, ErrConfigCastInvalid
}

func (wc *writableConfig) Reference() *Reference {
	return wc.origin.ref
}

func (wc *writableConfig) Set(key string, value interface{}) error {
	//if wc.ref != nil {
	//	return ErrConfigChangeNotAllowed
	//}
	//wc.m[key] = value
	return nil
}

func (wc *writableConfig) SetConfig(key string, config ConfigReadable) error {
	//if wc.ref != nil {
	//	return ErrConfigChangeNotAllowed
	//}
	//m := config.RawMap()
	//copied := make(map[string]interface{}, len(m))
	//for k, v := range m {
	//	copied[k] = v
	//}
	//wc.m[key] = copied
	return nil
}

func (wc *writableConfig) SubConfig(key string) (ConfigWritable, error) {
	//value, exists := wc.m[key]
	//if !exists {
	//	return nil, ErrConfigNotExists
	//}
	//m, ok := value.(map[string]interface{})
	//if !ok {
	//	return nil, ErrConfigCastInvalid
	//}
	//var path string
	//if len(wc.path) == 0 {
	//	path = key
	//} else {
	//	path = wc.path + "." + key
	//}
	//
	//newwc := writableConfig{
	//	readableConfig{
	//		path: path,
	//		m: m,
	//	},
	//}
	//
	//return &newwc, nil
	return nil, nil
}

func (wc *writableConfig) Merge(config ConfigReadable) error {
	//rc, err := wc.Shrink()
	//if err != nil {
	//	return err
	//}
	//
	//newkeys := config.Keys()


	return nil
}