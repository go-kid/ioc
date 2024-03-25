package ioc

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/registry"
)

func Register(cs ...interface{}) {
	for _, c := range cs {
		name, _ := component_definition.GetComponentName(c)
		registry.GlobalRegistry().RegisterSingleton(name, c)
	}
}
