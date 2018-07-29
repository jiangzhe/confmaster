package kv

// reference to part of remote config
type ConfigReference struct {
	Labels Labels `json:"labels,omitempty"`
	Path string `json:"path,omitempty"`
}

func NewConfigReference(labels Labels, path string) *ConfigReference {
	return &ConfigReference{
		Labels: labels,
		Path: path,
	}
}

func (cr *ConfigReference) Clone() *ConfigReference {
	return &ConfigReference{
		Labels: cr.Labels,
		Path: cr.Path,
	}
}