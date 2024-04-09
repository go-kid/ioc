package reflectx

import (
	"github.com/go-kid/ioc/util/fas"
	"path"
	"reflect"
)

func ZeroValue(p reflect.Type) any {
	return zeroValue(p, make(map[string]reflect.Value))
}

func zeroValue(p reflect.Type, cache map[string]reflect.Value) any {
	cacheKey := path.Join(fas.TernaryOpNil(p.Kind() == reflect.Pointer, func() string {
		return p.Elem().PkgPath()
	}, p.PkgPath), p.String())
	if cached, ok := cache[cacheKey]; ok {
		return cached.Interface()
	}
	var zero = reflect.New(p).Elem()

	switch p.Kind() {
	case reflect.Bool:
		zero.SetBool(false)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		zero.SetInt(0)
	case reflect.Float32, reflect.Float64:
		zero.SetFloat(0.0)
	case reflect.Array:
		arrType := reflect.ArrayOf(p.Len(), p.Elem())
		arr := reflect.New(arrType).Elem()
		if item := zeroValue(p.Elem(), cache); item != nil {
			for i := 0; i < p.Len(); i++ {
				arr.Index(i).Set(reflect.ValueOf(item))
			}
		}
		zero.Set(arr)
	case reflect.Slice:
		var slice = reflect.MakeSlice(p, 1, 1)
		if item := zeroValue(p.Elem(), cache); item != nil {
			slice.Index(0).Set(reflect.ValueOf(item))
		}
		zero.Set(slice)
	case reflect.Map:
		m := reflect.MakeMapWithSize(p, 1)
		keyType := zeroValue(p.Key(), cache)
		valueType := zeroValue(p.Elem(), cache)
		if keyType != nil && valueType != nil {
			m.SetMapIndex(reflect.ValueOf(keyType), reflect.ValueOf(valueType))
		}
		zero.Set(m)
	case reflect.Pointer:
		ptr := reflect.New(p.Elem())
		if actualVal := zeroValue(p.Elem(), cache); actualVal != nil {
			ptr.Elem().Set(reflect.ValueOf(actualVal))
		}
		zero.Set(ptr)
	case reflect.String:
		zero.SetString("string")
	case reflect.Struct:
		cache[cacheKey] = zero
		_ = ForEachFieldV2(p, zero, true, func(field reflect.StructField, value reflect.Value) error {
			if field.Tag.Get("json") == "-" ||
				field.Tag.Get("yaml") == "-" ||
				field.Tag.Get("mapstructure") == "-" {
				return nil
			}
			if zeroVal := zeroValue(field.Type, cache); zeroVal != nil {
				value.Set(reflect.ValueOf(zeroVal))
			}
			return nil
		})
	}
	cache[cacheKey] = zero
	return zero.Interface()
}
