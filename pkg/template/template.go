package template

type TemplateEncoding int

const (
	EncodingUTF8 TemplateEncoding = iota
	EncodingUTF16
	EncodingASCII
	//EncodingGB18030
)

type Template struct {
	Name string
	Filename string
	Raw []byte
	Encoding TemplateEncoding
}

type Store interface {
	Get(key string) (Template, bool)
	Set(key string, template Template)
}
