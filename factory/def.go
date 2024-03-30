package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/factory/support"
)

type MetaFunc func(m *component_definition.Meta) error

type Factory interface {
	GetRegisteredComponents() map[string]any
	AddDefinitionRegistryPostProcessors(processors ...processors.DefinitionRegistryPostProcessor)
	GetDefinitionRegistryPostProcessors() []processors.DefinitionRegistryPostProcessor
	SetRegistry(r support.SingletonRegistry)
	SetConfigure(c configure.Configure)
	AddInjectionRules(rules ...InjectionRule)
	PrepareComponents() error
	Refresh() error
	GetComponents(opts ...support.Option) ([]any, error)
	GetComponentByName(name string) (any, error)
	GetDefinitionRegistry() support.DefinitionRegistry
}

type InjectionRule interface {
	RuleName() string
	Priority() int
	Condition(d *component_definition.Node) bool
	Candidates(r support.DefinitionRegistry, d *component_definition.Node) ([]*component_definition.Meta, error)
}

type ComponentFactoryPostProcessor interface {
	PostProcessComponentFactory(factory Factory) error
}
