package kv

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"strings"
	"bufio"
	"strconv"
)

// formatter can format a resolved config to bytes
type Formatter interface {
	Format(config ResolvedConfigInterface) ([]byte, error)
}

func NewJsonFormatter(prefix int, indent int) Formatter {
	return &jsonFormatter{
		prefix: strings.Repeat(" ", prefix),
		indent: strings.Repeat(" ", indent),
	}
}

type jsonFormatter struct {
	prefix string
	indent string
}

func (jf *jsonFormatter) Format(config ResolvedConfigInterface) ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent(jf.prefix, jf.indent)
	err := encoder.Encode(config.ToMap())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func NewYamlFormatter(prefix int) Formatter {
	return &yamlFormatter{
		prefix: strings.Repeat(" ", prefix),
	}
}

// yaml formatter need to use special MapSlice to retain key order
type yamlFormatter struct {
	prefix string
}

func (yf *yamlFormatter) Format(config ResolvedConfigInterface) ([]byte, error) {
	bs, err := yaml.Marshal(config.ToMapSlice())
	if err != nil {
		return nil, err
	}
	if yf.prefix == "" {
		return bs, nil
	}
	scanner := bufio.NewScanner(bytes.NewReader(bs))
	buf := bytes.Buffer{}
	for scanner.Scan() {
		buf.WriteString(yf.prefix)
		buf.Write(scanner.Bytes())
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

func NewPropertiesFormatter() Formatter {
	return &propertiesFormatter{}
}

type propertiesFormatter struct {}

func (pf *propertiesFormatter) Format(config ResolvedConfigInterface) ([]byte, error) {
	//m := config.ToMap()
	//buf := new(bytes.Buffer)
	//for k, v := range m {
	//	formatProperties([]byte(k), v, buf)
	//}
	//return buf.Bytes(), nil
	co := config.ToConfig()
	buf := bytes.Buffer{}
	err := co.traverse(func(path string, value *Value) error {
		if value == nil {
			return nil
		}
		switch value.Type {
		case StringType:
			buf.Write([]byte(path))
			buf.WriteByte('=')
			buf.Write([]byte(value.RefValue.(string)))
			buf.WriteByte('\n')
		case BoolType:
			buf.Write([]byte(path))
			buf.WriteByte('=')
			buf.Write([]byte(strconv.FormatBool(value.RefValue.(bool))))
			buf.WriteByte('\n')
		case NumericType:
			buf.Write([]byte(path))
			buf.WriteByte('=')
			buf.Write([]byte(*value.RefValue.(*Number)))
			buf.WriteByte('\n')
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
