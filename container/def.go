package container

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
)

type SingletonRegistry interface {
	RegisterSingleton(singleton any)
	GetSingleton(name string) (any, error)
	ContainsSingleton(name string) bool
	GetSingletonNames() []string
	GetSingletonCount() int
}

type DefinitionRegistry interface {
	RegisterMeta(m *component_definition.Meta)
	GetMetas(opts ...Option) []*component_definition.Meta
	GetMetaByName(name string) *component_definition.Meta
	GetMetaOrRegister(name string, component any) *component_definition.Meta
}

type SingletonComponentRegistry interface {
	AddSingleton(name string, meta *component_definition.Meta)
	AddSingletonFactory(name string, method SingletonFactory)
	GetSingleton(name string, allowEarlyReference bool) (*component_definition.Meta, error)
	RemoveSingleton(name string)
	GetSingletonOrCreateByFactory(name string, factory SingletonFactory) (*component_definition.Meta, error)
	IsSingletonCurrentlyInCreation(name string) bool
}

type SingletonFactory interface {
	GetComponent() (*component_definition.Meta, error)
}

type FuncSingletonFactory func() (*component_definition.Meta, error)

func (d FuncSingletonFactory) GetComponent() (*component_definition.Meta, error) {
	return d()
}

type Factory interface {
	GetRegisteredComponents() map[string]any
	GetDefinitionRegistryPostProcessors() []DefinitionRegistryPostProcessor
	SetRegistry(r SingletonRegistry)
	SetConfigure(c configure.Configure)
	PrepareComponents() error
	Refresh() error
	GetComponents(opts ...Option) ([]any, error)
	GetComponentByName(name string) (any, error)
	GetConfigure() configure.Configure
	GetDefinitionRegistry() DefinitionRegistry
}

type ComponentFactoryPostProcessor interface {
	PostProcessComponentFactory(factory Factory) error
}

type DefinitionRegistryPostProcessor interface {
	PostProcessDefinitionRegistry(registry DefinitionRegistry, component any, componentName string) error
}

type ComponentPostProcessor interface {
	PostProcessBeforeInitialization(component any, componentName string) (any, error)
	PostProcessAfterInitialization(component any, componentName string) (any, error)
}

type InstantiationAwareComponentPostProcessor interface {
	ComponentPostProcessor
	PostProcessBeforeInstantiation(m *component_definition.Meta, componentName string) (any, error)
	PostProcessAfterInstantiation(component any, componentName string) (bool, error)
	PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error)
}

type SmartInstantiationAwareBeanPostProcessor interface {
	InstantiationAwareComponentPostProcessor
	GetEarlyBeanReference(component any, componentName string) (any, error)
}

type DestructionAwareComponentPostProcessor interface {
	ComponentPostProcessor
	PostProcessBeforeDestruction(component any, componentName string) error
	RequireDestruction(component any) bool
}
