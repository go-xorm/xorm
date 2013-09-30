package xorm

import (
	"reflect"
	"strings"
)

func indexNoCase(s, sep string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(sep))
}

func splitNoCase(s, sep string) []string {
	idx := indexNoCase(s, sep)
	if idx < 0 {
		return []string{s}
	}
	return strings.Split(s, s[idx:idx+len(sep)])
}

func splitNNoCase(s, sep string, n int) []string {
	idx := indexNoCase(s, sep)
	if idx < 0 {
		return []string{s}
	}
	return strings.SplitN(s, s[idx:idx+len(sep)], n)
}

func makeArray(elem string, count int) []string {
	res := make([]string, count)
	for i := 0; i < count; i++ {
		res[i] = elem
	}
	return res
}

func rType(bean interface{}) reflect.Type {
	sliceValue := reflect.Indirect(reflect.ValueOf(bean))
	return reflect.TypeOf(sliceValue.Interface())
}

func structName(v reflect.Type) string {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Name()
}
