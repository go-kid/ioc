package ioc

import (
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
)

func Register(cs ...interface{}) {
	registry.GlobalRegistry().Register(cs...)
}

func GetComponents(options ...registry.Option) []*meta.Meta {
	return registry.GlobalRegistry().GetComponents(options...)
}

func GetComponentByName(name string) *meta.Meta {
	return registry.GlobalRegistry().GetComponentByName(name)
}

func RemoveComponents(name string) {
	registry.GlobalRegistry().RemoveComponents(name)
}
