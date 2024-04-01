package instantiation_aware_component_post_processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/factory/support"
)

type dependencyNameAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	Registry support.DefinitionRegistry `wire:""`
}

func NewDependencyNameAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &dependencyNameAwarePostProcessors{}
}

func (d *dependencyNameAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (d *dependencyNameAwarePostProcessors) PostProcessProperties(properties []*component_definition.Node, component any, componentName string) ([]*component_definition.Node, error) {
	for _, property := range properties {
		if property.NodeType != component_definition.NodeTypeComponent {
			continue
		}
		fmt.Println(property.ID())
	}
	return properties, nil
}
