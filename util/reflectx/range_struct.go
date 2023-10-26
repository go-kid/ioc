package reflectx

import (
	"reflect"
)

type FieldAcceptor func(field reflect.StructField, value reflect.Value) error

func ForEachField(o interface{}, excludePrivateField bool, f FieldAcceptor) error {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	return ForEachFieldV2(t, v, excludePrivateField, f)
}

func ForEachFieldV2(t reflect.Type, v reflect.Value, excludePrivateField bool, f FieldAcceptor) error {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		if !excludePrivateField {
			if err := f(t.Field(i), v.Field(i)); err != nil {
				return err
			}
		} else if excludePrivateField && isPublicField(v.Field(i)) {
			if err := f(t.Field(i), v.Field(i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func isPublicField(field reflect.Value) bool {
	return field.CanSet()
}
