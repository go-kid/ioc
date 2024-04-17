package instantiation_aware_component_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/strconv2"
	"github.com/pkg/errors"
)

type valueAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	definition.LazyInitComponent
}

func NewValueAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &valueAwarePostProcessors{}
}

func (c *valueAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (c *valueAwarePostProcessors) Order() int {
	return PriorityOrderPopulateProperties
}

func (c *valueAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if prop.Tag != definition.ValueTag {
			continue
		}
		if prop.TagVal == "" {
			if prop.IsRequired() {
				return nil, errors.Errorf("value on '%s' is required", prop)
			}
			continue
		}
		parseVal, err := strconv2.ParseAny(prop.TagVal)
		if err != nil {
			return nil, errors.WithMessagef(err, "parse value on '%s' failed", prop)
		}
		err = prop.Unmarshall(parseVal)
		//err := reflectx.SetAnyValueFromString(prop.Type, prop.Value, prop.TagVal, c.hm)
		if err != nil {
			return nil, errors.WithMessagef(err, "populate on '%s' failed", prop)
		}
	}
	return nil, nil
}
