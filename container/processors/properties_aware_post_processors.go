package processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
	"github.com/pkg/errors"
)

type propertiesAwarePostProcessors struct {
	DefaultTagScanDefinitionRegistryPostProcessor
	DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	Configure configure.Configure
}

func NewPropertiesAwarePostProcessors() container.InstantiationAwareComponentPostProcessor {
	return &propertiesAwarePostProcessors{
		DefaultTagScanDefinitionRegistryPostProcessor: DefaultTagScanDefinitionRegistryPostProcessor{
			NodeType: component_definition.PropertyTypeConfiguration,
			Tag:      definition.PrefixTag,
			ExtractHandler: func(meta *component_definition.Meta, field *component_definition.Field) (tag, tagVal string, ok bool) {
				if configuration, infer := field.Value.Interface().(definition.ConfigurationProperties); infer {
					tagVal = configuration.Prefix()
					ok = true
				}
				return
			},
			Required: true,
		},
	}
}

func (c *propertiesAwarePostProcessors) PostProcessComponentFactory(factory container.Factory) error {
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
