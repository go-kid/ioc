package kid_ioc

import (
	"github.com/kidhat/kid-ioc/meta"
	"github.com/kidhat/kid-ioc/registry"
	"reflect"
)

var _registry = registry.NewRegistry()

func Register(cs ...interface{}) {
	_registry.Register(cs...)
}

func GetComponents() []*meta.Meta {
	return _registry.GetComponents()
}

func GetComponentByName(name string) *meta.Meta {
	return _registry.GetComponentByName(name)
}

func GetBeansByInterfaceType(typ reflect.Type) []*meta.Meta {
	return _registry.GetBeansByInterfaceType(typ)
}

func GetBeansByInterface(a interface{}) []*meta.Meta {
	return _registry.GetBeansByInterface(a)
}

func RemoveComponents(name string) {
	_registry.RemoveComponents(name)
}
