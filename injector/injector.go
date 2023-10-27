package injector

import "reflect"

type Injector interface {
	GetByName(name string) (reflect.Value, bool)
	GetOneByInterfaceType(typ reflect.Type) (reflect.Value, bool)
	GetsByInterfaceType(typ reflect.Type) []reflect.Value
	GetByFunc(funcName string) []reflect.Value
}
