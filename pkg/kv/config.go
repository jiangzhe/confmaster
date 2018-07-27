package kv

import "errors"

// a plain config used for key value read/write
type Config interface {

	// the path tells how to get current config from root
	// the root config has 0-length array as its root
	Path() []string

	// get value by key
	Get(key string) (interface{}, bool)

	// set value by given key
	// the method usually succeeds, but if the config is a depended one,
	// the modification will fail
	Set(Key string, value interface{}) error

	// get all keys
	Keys() []string

	// get all values
	Values() []interface{}

	// return a resolved config, all inner references will be resolved after this call
	// error out if any reference is not resolvable
	Resolve() (ResolvedConfig, error)

	// get and cast the value as config, useful for nested/structured key value settings
	// note the change on sub-config should be reflected to the parent
	GetConfig(key string) (Config, error)

	// set config by given key
	SetConfig(key string, config Config) error

	// whether the config is a reference to other app's config
	// if it's a reference, it should belong to some mapping of one dependency
	// note modification is disabled(return error) if it is reference
	Reference() bool
}

// resolved config is a complete self-contained config
// any reference config will be resolved to be local
type ResolvedConfig interface {
	Config

	ToMap() map[string]interface{}
}

//// reference to a remote config
//type ReferenceConfig interface {
//	Config
//
//	// convert the reference config to local config, it is convenient for
//	// user to replace the reference with local config
//	ToLocal() Config
//}

var (
	ErrConfigNotExists = errors.New("config not exists")
	ErrConfigChangeNotAllowed = errors.New("config change not allowed")
	ErrValueChangeNotAllowed = errors.New("value change not allowed")
)