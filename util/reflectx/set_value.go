package reflectx

import (
	"fmt"
	"github.com/go-kid/ioc/util/strconv2"
	"reflect"
	"strconv"
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

func SetAnyValue(t reflect.Type, value reflect.Value, val string) error {
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
	case reflect.Array, reflect.Slice:
		vals, err := strconv2.ParseStringSlice(val)
		if err != nil {
			return err
		}
		value.Set(reflect.MakeSlice(value.Type(), len(vals), len(vals)))
		for i, v := range vals {
			err := SetAnyValue(t.Elem(), value.Index(i), v)
			if err != nil {
				return err
			}
		}
	case reflect.Pointer:
		err := SetAnyValue(t.Elem(), value.Elem(), val)
		if err != nil {
			return err
		}
	case reflect.String:
		value.SetString(val)
	case reflect.Map:
		m, err := strconv2.ParseAnyMap(val)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(m))
	default:
		return fmt.Errorf("can not parse value %s", val)
	}
	return nil
}
