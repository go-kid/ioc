package ioc

import "reflect"

var _registry = NewRegistry()

func Register(cs ...interface{}) {
	_registry.Register(cs...)
}

func GetComponents() []*meta {
	return _registry.GetComponents()
}

func GetComponentByName(name string) *meta {
	return _registry.GetComponentByName(name)
}

func GetBeansByInterfaceType(typ reflect.Type) []*meta {
	return _registry.GetBeansByInterfaceType(typ)
}

func GetBeansByInterface(a interface{}) []*meta {
	return _registry.GetBeansByInterface(a)
}

func RemoveComponents(name string) {
	_registry.RemoveComponents(name)
}
