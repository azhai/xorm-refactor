package inspect

import (
	"fmt"
	"reflect"
	"strings"
)

func GetIndirectType(v interface{}) (rt reflect.Type) {
	var ok bool
	if rt, ok = v.(reflect.Type); !ok {
		rt = reflect.TypeOf(v)
	}
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	return
}

func GetFinalType(v interface{}) (rt reflect.Type) {
	rt = GetIndirectType(v)
	for {
		switch rt.Kind() {
		default:
			break
		case reflect.Ptr, reflect.Chan:
			rt = rt.Elem()
		case reflect.Array, reflect.Slice:
			rt = rt.Elem()
		case reflect.Map:
			kk := rt.Key().Kind()
			if kk == reflect.String || kk <= reflect.Float64 {
				rt = rt.Elem()
			} else {
				break
			}
		}
	}
	return
}

func GetColumns(v interface{}, alias string, cols []string) []string {
	rt := GetIndirectType(v)
	if rt.Kind() != reflect.Struct {
		return cols
	}
	for i := 0; i < rt.NumField(); i++ {
		t := rt.Field(i).Tag.Get("json")
		if t == "" || t == "-" {
			continue
		} else if strings.HasSuffix(t, "inline") {
			cols = GetColumns(rt.Field(i).Type, alias, cols)
		} else {
			if alias != "" {
				t = fmt.Sprintf("%s.%s", alias, t)
			}
			cols = append(cols, t)
		}
	}
	return cols
}
