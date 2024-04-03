package instantiation_aware_component_post_processors

import (
	"encoding/json"
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

type valueAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	definition.LazyInitComponent
	hm reflectx.Interceptor
}

func NewValueAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	var jsonUnmarshalHandler = func(_ reflect.Type, v reflect.Value, s string) error {
		return reflectx.SetValue(v, func(a any) error {
			return json.Unmarshal([]byte(s), a)
		})
	}
	return &valueAwarePostProcessors{
		hm: reflectx.Interceptor{
			reflect.Map:       jsonUnmarshalHandler,
			reflect.Slice:     jsonUnmarshalHandler,
			reflect.Struct:    jsonUnmarshalHandler,
			reflect.Interface: jsonUnmarshalHandler,
		},
	}
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
			continue
		}
		err := reflectx.SetAnyValueFromString(prop.Type, prop.Value, prop.TagVal, c.hm)
		if err != nil {
			return nil, fmt.Errorf("population 'value' value %s to %s failed: %v", prop.TagVal, prop.Type.String(), err)
		}
	}
	return nil, nil
}
