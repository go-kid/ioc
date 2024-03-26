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
	GetComponents(opts ...Option) []any
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
	ComponentInitialized(meta *component_definition.Meta)
	AddSingletonFactory(name string, method SingletonFactory)
	GetSingletonFactory(name string) (SingletonFactory, bool)
	EarlyExportComponent(m *component_definition.Meta)
	GetEarlyExportComponent(name string) (*component_definition.Meta, bool)
	GetComponentDefinitions(opts ...Option) []*component_definition.Meta
	GetComponentDefinitionByName(name string) (*component_definition.Meta, bool)
	GetComponent(name string) (*component_definition.Meta, error)
	BeforeSingletonCreation(name string)
	IsSingletonCurrentlyInCreation(name string) bool
	RemoveComponents(name string)
}

type SingletonFactory interface {
	GetComponent() (*component_definition.Meta, error)
}

type FuncSingletonFactory func() (*component_definition.Meta, error)

func (d FuncSingletonFactory) GetComponent() (*component_definition.Meta, error) {
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
	ComponentPostProcessor
	PostProcessBeforeInstantiation(p reflect.Type, componentName string) (any, error)
	PostProcessAfterInstantiation(component any, componentName string) (bool, error)
	PostProcessProperties(properties []*component_definition.Node, component any, componentName string) ([]*component_definition.Meta, error)
}

type ComponentInitializedPostProcessor interface {
	PostProcessBeforeInitialization(component any) error
	PostProcessAfterInitialization(component any) error
}

type DestructionAwareComponentPostProcessor interface {
	ComponentPostProcessor
	PostProcessBeforeDestruction(component any, componentName string) error
	RequireDestruction(component any) bool
}
