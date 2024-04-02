package factory

import (
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/factory/support"
)

type Factory interface {
	GetRegisteredComponents() map[string]any
	GetDefinitionRegistryPostProcessors() []processors.DefinitionRegistryPostProcessor
	SetRegistry(r support.SingletonRegistry)
	SetConfigure(c configure.Configure)
	PrepareComponents() error
	Refresh() error
	GetComponents(opts ...support.Option) ([]any, error)
	GetComponentByName(name string) (any, error)
	GetConfigure() configure.Configure
	GetDefinitionRegistry() support.DefinitionRegistry
}

type ComponentFactoryPostProcessor interface {
	PostProcessComponentFactory(factory Factory) error
}
