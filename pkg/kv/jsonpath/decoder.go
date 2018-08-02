package jsonpath

import (
	"encoding/json"
	"fmt"
	"io"
	"bytes"
	"strconv"
)

type jsonContext int

const (
	none jsonContext = iota
	objKey
	objValue
	arrValue
)

type KeyString string

type JsonPath []interface{}

// increment the index at the top of the stack (must be an array index)
func (p *JsonPath) inc() { (*p)[len(*p)-1] = (*p)[len(*p)-1].(int) + 1 }

// name the key at the top of the stack (must be an object key)
func (p *JsonPath) name(n string) { (*p)[len(*p)-1] = n }

func (p *JsonPath) push(n interface{}) { *p = append(*p, n) }

func (p *JsonPath) pop() { *p = (*p)[:len(*p)-1] }

func (p *JsonPath) inferContext() jsonContext {
	if len(*p) == 0 {
		return none
	}
	t := (*p)[len(*p)-1]
	switch t.(type) {
	case string:
		return objKey
	case int:
		return arrValue
	default:
		panic(fmt.Sprintf("Invalid stack type %T", t))
	}
}

func (p *JsonPath) String() string {
	if len(*p) == 0 {
		return ""
	}
	buf := bytes.Buffer{}
	for idx, key := range *p {
		switch k := key.(type) {
		case int:
			buf.WriteByte('[')
			buf.WriteString(strconv.Itoa(k))
			buf.WriteByte(']')
		case string:
			if idx > 0 {
				buf.WriteByte('.')
			}
			buf.WriteString(k)
		}
	}
	return buf.String()
}

type Decoder struct {
	json.Decoder
	context jsonContext
	path JsonPath
}

func NewDecoder(reader io.Reader) *Decoder {
	jsonDecoder := json.NewDecoder(reader)
	return &Decoder{
		Decoder: *jsonDecoder,
	}
}

func (d *Decoder) Decode(v interface{}) error {
	return d.Decoder.Decode(v)
}

func (d *Decoder) Path() JsonPath {
	path := make([]interface{}, len(d.path))
	copy(path, d.path)
	return path
}

// Token is equivalent to the Token() method on json.Decoder. The primary difference is that it distinguishes
// between strings that are keys and and strings that are values. String tokens that are object keys are returned as a
// KeyString rather than as a native string.
func (d *Decoder) Token() (json.Token, error) {
	t, err := d.Decoder.Token()
	if err != nil {
		return t, err
	}

	if t == nil {
		switch d.context {
		case objValue:
			d.context = objKey
		case arrValue:
			d.path.inc()
		}
		return t, err
	}

	switch t := t.(type) {
	case json.Delim:
		switch t {
		case json.Delim('{'):
			if d.context == arrValue {
				d.path.inc()
			}
			d.path.push("")
			d.context = objKey
		case json.Delim('}'):
			d.path.pop()
			d.context = d.path.inferContext()
		case json.Delim('['):
			if d.context == arrValue {
				d.path.inc()
			}
			d.path.push(-1)
			d.context = arrValue
		case json.Delim(']'):
			d.path.pop()
			d.context = d.path.inferContext()
		}
	case float64, json.Number, bool:
		switch d.context {
		case objValue:
			d.context = objKey
		case arrValue:
			d.path.inc()
		}
	case string:
		switch d.context {
		case objKey:
			d.path.name(t)
			d.context = objValue
			return KeyString(t), err
		case objValue:
			d.context = objKey
		case arrValue:
			d.path.inc()
		}
	}
	return t, err
}

type ScanAction func(JsonPath, json.Token) error

// scan the stream and invoke action on each token
func (d *Decoder) Scan(act ScanAction) (err error) {
	var token json.Token
	for {
		// advance the token position
		token, err = d.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if act != nil {
			err = act(d.Path(), token)
			if err != nil {
				return err
			}
		}
	}
	return
}