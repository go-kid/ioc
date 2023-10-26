package reflectx

import (
	"path"
	"reflect"
)

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

func TypeId(p reflect.Type) string {
	if p.Kind() == reflect.Pointer {
		p = p.Elem()
	}
	return path.Join(p.PkgPath(), p.Name())
}
