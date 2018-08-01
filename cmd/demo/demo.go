package main

import (
	"fmt"
	"io"
)

func main() {
	//s := "你好hello world"
	//fmt.Printf("len=%v\n", len(s))
	//fmt.Printf("lastindex(h)=%v\n", strings.LastIndex(s, "h"))
	//fmt.Printf("%v %v", s[6:], string([]byte(s)[6:]))

	var iface io.Reader = nil
	s, ok := iface.(io.Reader)
	if ok {
		fmt.Printf("ok %v", s)
	}
}
