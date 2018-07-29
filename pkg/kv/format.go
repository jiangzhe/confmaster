package kv

//
//// formatter can format a resolved config to bytes
//type Formatter interface {
//	Format(config ResolvedConfig) ([]byte, error)
//}
//
//func NewJsonFormatter() Formatter {
//	return &jsonFormatter{}
//}
//
//type jsonFormatter struct {}
//
//func (jf *jsonFormatter) Format(config ResolvedConfig) ([]byte, error) {
//	return json.MarshalIndent(config.ToMap(), "", "  ")
//}
//
//func NewYamlFormatter() Formatter {
//	return &yamlFormatter{}
//}
//
//type yamlFormatter struct {}
//
//func (yf *yamlFormatter) Format(config ResolvedConfig) ([]byte, error) {
//	return yaml.Marshal(config.ToMap())
//}
//
//type propertiesFormatter struct {}
//
//func (pf *propertiesFormatter) Format(config ResolvedConfig) ([]byte, error) {
//	m := config.ToMap()
//	buf := new(bytes.Buffer)
//	for k, v := range m {
//		formatProperties([]byte(k), v, buf)
//	}
//	return buf.Bytes(), nil
//}
//
//func formatProperties(path []byte, value interface{}, writer io.Writer) error {
//	m, ok := value.(map[string]interface{})
//	if ok {
//		for k, v := range m {
//			p := make([]byte, 0, len(path) + len(k) + 1)
//			p = append(p, []byte(path)...)
//			p = append(p, '.')
//			p = append(p, []byte(k)...)
//			if err := formatProperties(p, v, writer); err != nil {
//				return err
//			}
//		}
//		return nil
//	}
//	a, ok := value.([]interface{})
//	if ok {
//		for i, v := range a {
//			p := append(path, []byte(fmt.Sprintf(".[%v]", i))...)
//			if err := formatProperties(p, v, writer); err != nil {
//				return err
//			}
//		}
//		return nil
//	}
//	if _, err := writer.Write([]byte(fmt.Sprintf("%v=%v\n", string(path), value))); err != nil {
//		return err
//	}
//	return nil
//}
