package factory

import (
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
)

type MetaFunc func(m *meta.Meta) error

type Factory interface {
	SetIfNilPostInitFunc(fn MetaFunc)
	Initialize(r registry.Registry, i injector.Injector, m *meta.Meta) error
}
