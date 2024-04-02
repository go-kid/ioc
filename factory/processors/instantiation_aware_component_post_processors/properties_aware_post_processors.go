package instantiation_aware_component_post_processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/util/reflectx"
)

type propertiesAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	Configure configure.Configure
}

func NewPropertiesAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &propertiesAwarePostProcessors{}
}

func (c *propertiesAwarePostProcessors) PostProcessComponentFactory(factory factory.Factory) error {
	c.Configure = factory.GetConfigure()
	return nil
}

func (c *propertiesAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (c *propertiesAwarePostProcessors) Order() int {
	return 2
}

func (c *propertiesAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if prop.Tag != definition.PropTag {
			continue
		}
		if c.Configure.Get(prop.TagVal) == nil {
			continue
		}
		err := reflectx.SetValue(prop.Value, func(a any) error {
			return c.Configure.Unmarshall(prop.TagVal, a)
		})
		if err != nil {
			return nil, fmt.Errorf("population 'prop' value %s to %s failed: %v", prop.TagVal, prop.Type.String(), err)
		}
	}
	return nil, nil
}
