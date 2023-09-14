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

func WalkField(o interface{}, f FieldAcceptor) error {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	return WalkFieldV2(t, v, f)
}

func WalkFieldV2(t reflect.Type, v reflect.Value, f FieldAcceptor) error {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	return ForEachFieldV2(t, v, false, func(field reflect.StructField, value reflect.Value) error {
		err := walkField(field, value, f)
		if err != nil {
			return err
		}
		return nil
	})
}

func walkField(field reflect.StructField, value reflect.Value, f FieldAcceptor) error {
	err := f(field, value)
	if err != nil {
		return err
	}
	t := field.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		value = value.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	if t.NumField() < 1 {
		return nil
	}
	for i := 0; i < t.NumField(); i++ {
		innerField := t.Field(i)
		innerValue := value.Field(i)
		err := walkField(innerField, innerValue, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func isPublicField(field reflect.Value) bool {
	return field.CanSet()
}

func TryCallMethod(b interface{}, methodName string, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	handleMethod := reflect.ValueOf(b).MethodByName(methodName)
	if handleMethod.IsValid() {
		values := handleMethod.Call(inputs)
		return values
	}
	return nil
}

func IsImplement(instance interface{}, _interface interface{}) bool {
	return IsTypeImplement(reflect.TypeOf(instance), _interface)
}

func IsTypeImplement(typ reflect.Type, _interface interface{}) bool {
	return typ.Implements(reflect.TypeOf(_interface).Elem())
}

func Values2Interfaces(values []reflect.Value) []interface{} {
	var result []interface{}
	for i := range values {
		result = append(result, values[i].Interface())
	}
	return result
}

func Interfaces2Values(o []interface{}) []reflect.Value {
	var values []reflect.Value
	for i := range o {
		values = append(values, reflect.ValueOf(o[i]))
	}
	return values
}

func New(t reflect.Type) reflect.Value {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t)
}

func Set(dst, src reflect.Value) {
	if dst.Type().Kind() == reflect.Ptr {
		dst.Set(src)
	} else {
		dst.Set(src.Elem())
	}
}
