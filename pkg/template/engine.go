package template

import "confmaster/pkg/kv"

// engine will render the template with given config
type Engine interface {
	Render(*Template, kv.Config) ([]byte, error)
}
