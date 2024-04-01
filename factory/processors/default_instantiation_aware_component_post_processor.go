package processors

import (
	"github.com/go-kid/ioc/component_definition"
)

type DefaultInstantiationAwareComponentPostProcessor struct {
	DefaultComponentPostProcessor
}

func (d *DefaultInstantiationAwareComponentPostProcessor) PostProcessBeforeInstantiation(m *component_definition.Meta, componentName string) (any, error) {
	return nil, nil
}

func (d *DefaultInstantiationAwareComponentPostProcessor) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return false, nil
}

func (d *DefaultInstantiationAwareComponentPostProcessor) PostProcessProperties(properties []*component_definition.Node, component any, componentName string) ([]*component_definition.Node, error) {
	return nil, nil
}
