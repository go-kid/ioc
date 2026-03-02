package ioc

import (
	"fmt"
	"reflect"
)

// Provide registers a typed constructor with compile-time return type verification.
// T is the expected component type. The constructor must be a function whose first
// return value is *T (if T is a concrete type) or implements T (if T is an interface).
// Panics at registration time if the types do not match.
func Provide[T any](constructor any) {
	ct := reflect.TypeOf(constructor)
	if ct == nil || ct.Kind() != reflect.Func {
		panic("ioc.Provide: constructor must be a function")
	}
	if ct.NumOut() < 1 || ct.NumOut() > 2 {
		panic(fmt.Sprintf("ioc.Provide: constructor must return 1 or 2 values, got %d", ct.NumOut()))
	}

	targetType := reflect.TypeOf((*T)(nil)).Elem()
	returnType := ct.Out(0)

	if targetType.Kind() == reflect.Interface {
		if !returnType.Implements(targetType) {
			panic(fmt.Sprintf("ioc.Provide: constructor return type %s does not implement %s", returnType, targetType))
		}
	} else {
		expectedPtr := reflect.PointerTo(targetType)
		if returnType != expectedPtr {
			panic(fmt.Sprintf("ioc.Provide: constructor return type %s does not match *%s", returnType, targetType))
		}
	}

	if ct.NumOut() == 2 {
		errType := reflect.TypeOf((*error)(nil)).Elem()
		if !ct.Out(1).Implements(errType) {
			panic(fmt.Sprintf("ioc.Provide: constructor second return value must be error, got %s", ct.Out(1)))
		}
	}

	Register(constructor)
}
