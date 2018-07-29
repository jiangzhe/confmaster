package kv

type Reference struct {
	Namespace Namespace `json:"namespace,omitempty"`
	Labels Labels `json:"labels,omitempty"`
	Path string `json:"path,omitempty"`
}

type ResolvedConfig interface {
	ToMap() map[string]interface{}

	ToReadable() ConfigReadable

	ToWritable() ConfigWritable
}

type Resolver interface {
	// return a resolved config, all inner references will be resolved after this call
	// error out if any reference is not resolvable
	Resolve(ConfigReadable) (ResolvedConfig, error)
}

