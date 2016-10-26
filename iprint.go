// Used for print everything as json.Marshal
// Created by simplejia [9/2016]
package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
)

func val2val(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}
	return val
}

func obj2json(v interface{}) (ret interface{}) {
	if v == nil {
		return nil
	}

	typ := reflect.TypeOf(v)
	if typ.Implements(reflect.TypeOf((*json.Marshaler)(nil)).Elem()) {
		ret = v
		return
	}
	if typ.Implements(reflect.TypeOf((*error)(nil)).Elem()) || typ.Implements(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()) {
		ret = fmt.Sprintf("%v", v)
		return
	}

	val := val2val(reflect.ValueOf(v))
	if !val.IsValid() {
		return
	}

	typ = val.Type()

	switch typ.Kind() {
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			ret = fmt.Sprintf("%s", v)
		} else {
			s := make([]interface{}, 0)
			for i := 0; i < val.Len(); i++ {
				s = append(s, obj2json(val.Index(i).Interface()))
			}
			ret = s
		}
	case reflect.Map:
		m := make(map[string]interface{})
		for _, mk := range val.MapKeys() {
			key := fmt.Sprintf("%v", val2val(mk).Interface())
			m[key] = obj2json(val.MapIndex(mk).Interface())
		}
		ret = m
	case reflect.Struct:
		m := make(map[string]interface{})
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if f.PkgPath != "" {
				continue
			}

			tag := f.Tag.Get("json")
			if tag == "-" {
				continue
			}

			name := f.Name
			if tag != "" {
				if n := strings.Split(tag, ",")[0]; n != "" {
					name = n
				}
			}
			fv := val.FieldByIndex(f.Index)
			m[name] = obj2json(fv.Interface())
		}
		ret = m
	default:
		ret = val.Interface()
	}

	return
}

func IprintD(a ...interface{}) {
	for _, v := range a {
		bs, err := json.Marshal(obj2json(v))
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(string(bs))
	}
	return
}

func Iprint(a ...interface{}) string {
	buf := new(bytes.Buffer)
	for pos, v := range a {
		if pos > 0 {
			buf.WriteByte('\n')
		}
		bs, err := json.Marshal(obj2json(v))
		if err != nil {
			buf.WriteString(err.Error())
		} else {
			buf.Write(bs)
		}
	}
	return buf.String()
}
