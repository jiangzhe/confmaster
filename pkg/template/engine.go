package template

import "confmaster/pkg/kv"

// render result, with additional render information
type RenderResult struct {
	Origin *Template
	Bytes []byte
	Err error
}

// engine will render the template with given config
type Engine interface {
	Render(*Template, kv.Config) RenderResult
}
