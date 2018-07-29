package template

// engine will render the template with given config
type Engine interface {
	Render(*Template, map[string]interface{}) ([]byte, error)
}
