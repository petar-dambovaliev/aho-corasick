package aho_corasick

import (
	"reflect"
	"unsafe"
)

func unsafeBytes(s string) []byte {
	str := (*reflect.StringHeader)(unsafe.Pointer(&s))
	slice := reflect.SliceHeader{Data: str.Data, Len: str.Len, Cap: str.Len}
	return *(*[]byte)(unsafe.Pointer(&slice))
}
