package instantiation_aware_component_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/pkg/errors"
)

type propertiesAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	definition.LazyInitComponent
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
	return PriorityOrderPopulateProperties
}

func (c *propertiesAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if prop.Tag != definition.PrefixTag {
			continue
		}

		configValue := c.Configure.Get(prop.TagVal)
		prop.SetConfiguration(prop.TagVal, configValue)
		if configValue == nil {
			if prop.IsRequired() {
				return nil, errors.Errorf("config value on '%s' is required", prop)
			}
			continue
		}
		err := prop.Unmarshall(configValue)
		if err != nil {
			return nil, errors.WithMessagef(err, "populate config value on '%s' failed", prop)
		}
	}
	return nil, nil
}
