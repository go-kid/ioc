package ioc

import (
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
)

var _registry = registry.NewRegistry()

func Register(cs ...interface{}) {
	_registry.Register(cs...)
}

func GetComponents(options ...registry.Option) []*meta.Meta {
	return _registry.GetComponents(options...)
}

func GetComponentByName(name string) *meta.Meta {
	return _registry.GetComponentByName(name)
}

func RemoveComponents(name string) {
	_registry.RemoveComponents(name)
}

func GlobalRegistry() registry.Registry {
	return _registry
}
