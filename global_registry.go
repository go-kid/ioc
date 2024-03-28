package ioc

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/registry"
)

func Register(cs ...interface{}) {
	registry.GlobalRegistry().Register(cs...)
}

func GetComponents(options ...registry.Option) []*component_definition.Meta {
	return registry.GlobalRegistry().GetComponents(options...)
}

func GetComponentByName(name string) *component_definition.Meta {
	return registry.GlobalRegistry().GetComponentByName(name)
}

func RemoveComponents(name string) {
	registry.GlobalRegistry().RemoveComponents(name)
}
