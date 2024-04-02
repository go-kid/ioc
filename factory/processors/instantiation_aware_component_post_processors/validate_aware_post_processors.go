package instantiation_aware_component_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-playground/validator/v10"
	"reflect"
)

const (
	ArgValidate component_definition.ArgType = "validate"
)

type validateAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	v *validator.Validate
}

func NewValidateAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &validateAwarePostProcessors{
		v: validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (c *validateAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (c *validateAwarePostProcessors) Order() int {
	return OrderValidate
}

func (c *validateAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if prop.PropertyType != component_definition.PropertyTypeConfiguration {
			continue
		}
		if ts, ok := prop.Args().Find(ArgValidate); ok {
			var p = prop.Type
			if p.Kind() == reflect.Pointer {
				p = p.Elem()
			}
			if p.Kind() == reflect.Struct {
				return nil, c.v.Struct(prop.Value.Interface())
			} else {
				for _, t := range ts {
					err := c.v.Var(prop.Value.Interface(), t)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return nil, nil
}
