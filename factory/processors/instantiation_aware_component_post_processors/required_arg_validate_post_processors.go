package instantiation_aware_component_post_processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/factory/processors"
)

type requiredArgValidatePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
}

func NewRequiredArgValidatePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &requiredArgValidatePostProcessors{}
}

func (c *requiredArgValidatePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (c *requiredArgValidatePostProcessors) Order() int {
	return 0
}

func (c *requiredArgValidatePostProcessors) PostProcessProperties(properties []*component_definition.Node, component any, componentName string) ([]*component_definition.Node, error) {
	for _, prop := range properties {
		if prop.NodeType != component_definition.NodeTypeConfiguration {
			continue
		}
		if prop.Value.IsZero() {
			if prop.Args().Has(component_definition.ArgRequired, "true") {
				return nil, fmt.Errorf("properties %s is required", prop.ID())
			}
			continue
		}
	}
	return properties, nil
}
