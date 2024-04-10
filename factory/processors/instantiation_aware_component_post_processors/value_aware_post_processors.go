package instantiation_aware_component_post_processors

import (
	"encoding/json"
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/ioc/util/strconv2"
	"reflect"
)

type valueAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	definition.LazyInitComponent
	hm reflectx.Interceptor
}

func NewValueAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &valueAwarePostProcessors{
		hm: reflectx.Interceptor{
			//reflect.Struct: reflectx.JsonUnmarshallHandler,
			reflect.Struct: unmarshallStructHandler,
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
			if prop.Args().Has(component_definition.ArgRequired, "true") {
				return nil, fmt.Errorf("value %s is required", prop.ID())
			}
			continue
		}
		err := reflectx.SetAnyValueFromString(prop.Type, prop.Value, prop.TagVal, c.hm)
		if err != nil {
			return nil, fmt.Errorf("population 'value' value %s to %s failed: %v", prop.TagVal, prop.Type.String(), err)
		}
	}
	return nil, nil
}

var unmarshallStructHandler reflectx.SetValueHandler = func(r reflect.Type, v reflect.Value, s string) error {
	return reflectx.SetValue(v, func(a any) error {
		var jsonBytes []byte
		if bytes := []byte(s); json.Valid(bytes) {
			jsonBytes = bytes
		} else {
			m, err := strconv2.ParseAnyMap(s)
			if err != nil {
				return err
			}
			jsonBytes, err = json.Marshal(m)
			if err != nil {
				return err
			}
		}
		err := json.Unmarshal(jsonBytes, a)
		if err != nil {
			return fmt.Errorf("unmarshall json %s to type '%s' failed: %v", s, r.String(), err)
		}
		return nil
	})
}
