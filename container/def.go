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

// ComponentFactoryPostProcessor Used for Component to get Factory
type ComponentFactoryPostProcessor interface {
	PostProcessComponentFactory(factory Factory) error
}

// DefinitionRegistryPostProcessor Used to add custom Component Meta parsing capabilities
type DefinitionRegistryPostProcessor interface {
	PostProcessDefinitionRegistry(registry DefinitionRegistry, component any, componentName string) error
}

// ComponentPostProcessor is an interface that provides extension points to modify components during their lifecycle in the IoC container
type ComponentPostProcessor interface {
	// PostProcessBeforeInitialization called before InitializingComponent.AfterPropertiesSet and InitializeComponent.Init
	PostProcessBeforeInitialization(component any, componentName string) (any, error)
	// PostProcessAfterInitialization called after InitializingComponent.AfterPropertiesSet and InitializeComponent.Init
	PostProcessAfterInitialization(component any, componentName string) (any, error)
}

// InstantiationAwareComponentPostProcessor is an extension of ComponentPostProcessor
// that provides additional callbacks for intercepting component instantiation and dependency injection.
// It allows fine-grained control over the component lifecycle, particularly before instantiation and after property population (dependency injection).
type InstantiationAwareComponentPostProcessor interface {
	ComponentPostProcessor
	// PostProcessBeforeInstantiation Return a proxy instead of the real bean
	PostProcessBeforeInstantiation(m *component_definition.Meta, componentName string) (any, error)
	// PostProcessAfterInstantiation Return false to skip injection (PostProcessProperties) for certain components.
	PostProcessAfterInstantiation(component any, componentName string) (bool, error)
	// PostProcessProperties inject and modify the properties value for certain components
	PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error)
}

// SmartInstantiationAwareBeanPostProcessor is an advanced extension of InstantiationAwareComponentPostProcessor
// that provides additional capabilities for predicting component types,
// This interface is primarily used for:
//
// - Circular dependency handling (exposing early bean references)
// - Proxy creation optimization
type SmartInstantiationAwareBeanPostProcessor interface {
	InstantiationAwareComponentPostProcessor
	GetEarlyBeanReference(component any, componentName string) (any, error)
}

type DestructionAwareComponentPostProcessor interface {
	ComponentPostProcessor
	PostProcessBeforeDestruction(component any, componentName string) error
	RequireDestruction(component any) bool
}
