package reflectx

import (
	"github.com/samber/lo"
	"path"
	"reflect"
)

func TryCallMethod(b interface{}, methodName string, args ...any) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	var handleMethod reflect.Value
	switch b.(type) {
	case reflect.Value:
		handleMethod = b.(reflect.Value).MethodByName(methodName)
	default:
		handleMethod = reflect.ValueOf(b).MethodByName(methodName)
	}
	if handleMethod.IsValid() {
		values := handleMethod.Call(inputs)
		return values
	}
	return nil
}

func IsImplement(instance any, _interface any) bool {
	return IsTypeImplement(reflect.TypeOf(instance), _interface)
}

func IsTypeImplement(typ reflect.Type, _interface any) bool {
	return typ.Implements(reflect.TypeOf(_interface).Elem())
}

func Values2Interfaces(values []reflect.Value) []any {
	return lo.Map(values, func(v reflect.Value, _ int) any { return v.Interface() })
}

func Interfaces2Values(o []any) []reflect.Value {
	return lo.Map(o, func(x any, _ int) reflect.Value { return reflect.ValueOf(x) })
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

func Id(c any) string {
	if c == nil {
		return "<nil>"
	}
	return TypeId(reflect.TypeOf(c))
}

func TypeId(p reflect.Type) string {
	if p.Kind() == reflect.Pointer {
		p = p.Elem()
	}
	if p.Name() == "" {
		return p.String()
	}
	return path.Join(p.PkgPath(), p.Name())
}
