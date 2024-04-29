package processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/strconv2"
	"github.com/go-kid/strings2"
	"github.com/pkg/errors"
)

type valueAwarePostProcessors struct {
	DefaultTagScanDefinitionRegistryPostProcessor
	DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
}

func NewValueAwarePostProcessors() container.InstantiationAwareComponentPostProcessor {
	return &valueAwarePostProcessors{
		DefaultTagScanDefinitionRegistryPostProcessor: DefaultTagScanDefinitionRegistryPostProcessor{
			NodeType: component_definition.PropertyTypeConfiguration,
			Tag:      definition.ValueTag,
			ExtractHandler: func(meta *component_definition.Meta, field *component_definition.Field) (tag, tagVal string, ok bool) {
				if tagVal, ok = field.StructField.Tag.Lookup(definition.PropTag); ok {
					var argstr string
					i := strings2.IndexSkipBlocks(tagVal, ",")
					if i != -1 {
						tagVal, argstr = tagVal[:i], tagVal[i:]
					}
					tagVal = fmt.Sprintf("${%s}%s", tagVal, argstr)
				}
				return
			},
			Required: true,
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
