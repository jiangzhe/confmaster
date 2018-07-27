package kv

type FormatType int

const (
	FormatJson FormatType = iota
	FormatYaml
	FormatProperties
	FormatXML
)

type Formatter interface {
	Format(config Config) ([]byte, error)
}


