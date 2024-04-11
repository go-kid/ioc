package processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/factory/support"
)

type DefinitionRegistryPostProcessor interface {
	PostProcessDefinitionRegistry(registry support.DefinitionRegistry, component any, componentName string) error
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
