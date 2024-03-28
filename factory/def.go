package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"reflect"
)

type MetaFunc func(m *component_definition.Meta) error

type Factory interface {
	SetRegistry(r SingletonRegistry)
	SetConfigure(c configure.Configure)
	AddInjectionRules(rules ...InjectionRule)
	PrepareComponents() error
	Refresh() error
	GetComponents(opts ...Option) ([]any, error)
	GetComponentByName(name string) (any, error)
	GetDefinitionRegistry() DefinitionRegistry
}

type InjectionRule interface {
	RuleName() string
	Priority() int
	Condition(d *component_definition.Node) bool
	Candidates(r DefinitionRegistry, d *component_definition.Node) ([]*component_definition.Meta, error)
}

type SingletonRegistry interface {
	RegisterSingleton(name string, singleton any)
	GetSingleton(name string) (any, error)
	ContainsSingleton(name string) bool
	GetSingletonNames() []string
	GetSingletonCount() int
}

type DefinitionRegistry interface {
	RegisterMeta(m *component_definition.Meta)
	GetMetas(opts ...Option) []*component_definition.Meta
	GetMetaByName(name string) *component_definition.Meta
}

type SingletonComponentRegistry interface {
	AddSingleton(name string, meta any)
	AddSingletonFactory(name string, method SingletonFactory)
	GetSingletonByFactory(name string, factory SingletonFactory) (any, error)
	GetSingleton(name string) (any, error)
	BeforeSingletonCreation(name string) error
	IsSingletonCurrentlyInCreation(name string) bool
	RemoveComponents(name string)
}

type SingletonFactory interface {
	GetComponent() (any, error)
}

type FuncSingletonFactory func() (any, error)

func (d FuncSingletonFactory) GetComponent() (any, error) {
	return d()
}

type DefinitionRegistryPostProcessor interface {
	PostProcessDefinitionRegistry(registry DefinitionRegistry, component any, componentName string) error
}

type ComponentFactoryPostProcessor interface {
	PostProcessComponentFactory(factory Factory) error
}

type ComponentPostProcessor interface {
	PostProcessBeforeInitialization(component any, componentName string) (any, error)
	PostProcessAfterInitialization(component any, componentName string) (any, error)
}

type InstantiationAwareComponentPostProcessor interface {
	PostProcessBeforeInstantiation(p reflect.Type, componentName string) (any, error)
	PostProcessAfterInstantiation(component any, componentName string) (bool, error)
	PostProcessProperties(properties []*component_definition.Node, component any, componentName string) ([]*component_definition.Meta, error)
}

type ComponentInitializedPostProcessor interface {
	PostProcessBeforeInitialization(component any) error
	PostProcessAfterInitialization(component any) error
}

type DestructionAwareComponentPostProcessor interface {
	PostProcessBeforeDestruction(component any, componentName string) error
	RequireDestruction(component any) bool
}
