package reflectx

import (
	"encoding/json"
	"fmt"
	"github.com/go-kid/ioc/util/strconv2"
	"reflect"
	"strconv"
	"strings"
)

var bitSizeMap = map[reflect.Kind]int{
	reflect.Int:     64,
	reflect.Int8:    8,
	reflect.Int16:   16,
	reflect.Int32:   32,
	reflect.Int64:   64,
	reflect.Uint:    64,
	reflect.Uint8:   8,
	reflect.Uint16:  16,
	reflect.Uint32:  32,
	reflect.Uint64:  64,
	reflect.Float32: 32,
	reflect.Float64: 64,
}

type (
	SetValueHandler func(reflect.Type, reflect.Value, string) error
	Interceptor     map[reflect.Kind]SetValueHandler
)

var JsonUnmarshallHandler SetValueHandler = func(r reflect.Type, v reflect.Value, s string) error {
	return SetValue(v, func(a any) error {
		err := json.Unmarshal([]byte(s), a)
		if err != nil {
			return fmt.Errorf("unmarshall json %s to type '%s' failed: %v", s, r.String(), err)
		}
		return nil
	})
}

func SetAnyValueFromString(t reflect.Type, value reflect.Value, val string, itps ...Interceptor) error {
	for _, hm := range itps {
		if h, ok := hm[t.Kind()]; ok {
			return h(t, value, val)
		}
	}
	switch kind := t.Kind(); kind {
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		value.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, bitSizeMap[kind])
		if err != nil {
			return err
		}
		value.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(val, 10, bitSizeMap[kind])
		if err != nil {
			return err
		}
		value.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, bitSizeMap[kind])
		if err != nil {
			return err
		}
		value.SetFloat(f)
	case reflect.Slice:
		vals, err := strconv2.ParseStringSlice(val)
		if err != nil {
			return err
		}
		value.Set(reflect.MakeSlice(t, len(vals), len(vals)))
		for i, v := range vals {
			err := SetAnyValueFromString(t.Elem(), value.Index(i), v, itps...)
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		vals, err := strconv2.ParseStringSlice(val)
		if err != nil {
			return err
		}
		if t.Len() != len(vals) {
			return fmt.Errorf("array length not match, want: %s, actual: %s", t.String(), vals)
		}
		for i, v := range vals {
			err := SetAnyValueFromString(t.Elem(), value.Index(i), v, itps...)
			if err != nil {
				return err
			}
		}
	case reflect.Pointer:
		if value.IsNil() {
			value.Set(reflect.New(t.Elem()))
		}
		err := SetAnyValueFromString(t.Elem(), value.Elem(), val, itps...)
		if err != nil {
			return err
		}
	case reflect.String:
		val = strings.Trim(val, "\"")
		value.SetString(val)
	case reflect.Interface:
		a, err := strconv2.ParseAny(val)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(a))
	case reflect.Map:
		m, err := strconv2.ParseAnyMap(val)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(m))
	default:
		return fmt.Errorf("not supported to set value %s as type %s", val, t.String())
	}
	return nil
}

func SetValue(value reflect.Value, setter func(a any) error) error {
	var fieldType = value.Type()
	var isPtrType = false
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
		isPtrType = true
	}
	var val = reflect.New(fieldType)
	err := setter(val.Interface())
	if err != nil {
		return err
	}
	if isPtrType {
		value.Set(val)
	} else {
		value.Set(val.Elem())
	}
	return nil
}
