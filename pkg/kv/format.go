package kv

import (
	"bytes"
	"fmt"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io"
	"strings"
	"bufio"
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

type yamlFormatter struct {
	prefix string
}

func (yf *yamlFormatter) Format(config ResolvedConfigInterface) ([]byte, error) {
	bs, err := yaml.Marshal(config.ToMap())
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
	m := config.ToMap()
	buf := new(bytes.Buffer)
	for k, v := range m {
		formatProperties([]byte(k), v, buf)
	}
	return buf.Bytes(), nil
}

func formatProperties(path []byte, value interface{}, writer io.Writer) error {
	m, ok := value.(map[string]interface{})
	if ok {
		for k, v := range m {
			p := make([]byte, 0, len(path) + len(k) + 1)
			p = append(p, []byte(path)...)
			p = append(p, '.')
			p = append(p, []byte(k)...)
			if err := formatProperties(p, v, writer); err != nil {
				return err
			}
		}
		return nil
	}
	a, ok := value.([]interface{})
	if ok {
		for i, v := range a {
			p := append(path, []byte(fmt.Sprintf(".[%v]", i))...)
			if err := formatProperties(p, v, writer); err != nil {
				return err
			}
		}
		return nil
	}
	if _, err := writer.Write([]byte(fmt.Sprintf("%v=%v\n", string(path), value))); err != nil {
		return err
	}
	return nil
}
