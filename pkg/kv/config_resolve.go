package kv

type Resolver interface {
	// return a resolved config, all inner references will be resolved after this call
	// error out if any reference is not resolvable
	Resolve(ConfigInterface) (ResolvedConfigInterface, error)
}

type ResolvedConfigInterface interface {

	ToConfig() ConfigInterface

	ToMap() map[string]interface{}

	Format(formatter Formatter) ([]byte, error)
}

// any config object without references can be treated as
// a ResolvedConfigObject
type ResolvedConfigObject struct {
	m *map[string]*Value
}

func (rco *ResolvedConfigObject) ToConfig() ConfigInterface {
	return &ConfigObject{
		m: rco.m,
		refs: make(map[string]*ConfigReference),
	}
}

func (rco *ResolvedConfigObject) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range *rco.m {
		m[k] = v.Unwrap()
	}
	return m
}

func (rco *ResolvedConfigObject) Format(formatter Formatter) ([]byte, error) {
	return formatter.Format(rco)
}
