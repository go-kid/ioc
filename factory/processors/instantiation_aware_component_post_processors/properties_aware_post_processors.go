package instantiation_aware_component_post_processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/mitchellh/mapstructure"
)

const (
	defaultUnmarshalTagName = "yaml"
	unmarshallConfigTagName = "app.config.unmarshallTagName"
	unmarshallArgTimeLayout = "timeLayout"
)

type propertiesAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	definition.LazyInitComponent
	Configure        configure.Configure
	unmarshalTagName string
}

func NewPropertiesAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &propertiesAwarePostProcessors{}
}

func (c *propertiesAwarePostProcessors) PostProcessComponentFactory(factory factory.Factory) error {
	c.Configure = factory.GetConfigure()
	if tagName := c.Configure.Get(unmarshallConfigTagName); tagName != nil {
		if s, ok := tagName.(string); ok {
			c.unmarshalTagName = s
			syslog.Pref("PropAwarePostProcessor").Debugf("configure unmarshall tag name change to use '%s'", s)
		}
	} else {
		c.unmarshalTagName = defaultUnmarshalTagName
	}
	syslog.Pref("PropAwarePostProcessor").Tracef("configure unmarshall tag name use '%s'", c.unmarshalTagName)
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
		if prop.Tag != definition.PropTag {
			continue
		}
		configValue := c.Configure.Get(prop.TagVal)
		if configValue == nil {
			if prop.Args().Has(component_definition.ArgRequired, "true") {
				return nil, fmt.Errorf("config value on '%s' is required", prop.ID())
			}
			continue
		}
		err := reflectx.SetValue(prop.Value, func(a any) error {
			config := newDecodeConfig(a)
			config.TagName = c.unmarshalTagName
			if args, ok := prop.Args().Find(unmarshallArgTimeLayout); ok {
				config.DecodeHook = mapstructure.ComposeDecodeHookFunc(config.DecodeHook, mapstructure.StringToTimeHookFunc(args[0]))
			}
			decoder, err := mapstructure.NewDecoder(config)
			if err != nil {
				return fmt.Errorf("create mapstructure decoder error: %v", err)
			}
			err = decoder.Decode(configValue)
			if err != nil {
				return fmt.Errorf("mapstructure decode %+v error: %v", configValue, err)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("populate on '%s' to %s failed: %v", prop.ID(), prop.Type.String(), err)
		}
	}
	return nil, nil
}

func newDecodeConfig(v any) *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
		ErrorUnused:          false,
		ErrorUnset:           false,
		ZeroFields:           false,
		WeaklyTypedInput:     true,
		Squash:               false,
		Metadata:             nil,
		Result:               v,
		TagName:              "",
		IgnoreUntaggedFields: false,
		MatchName:            nil,
	}
}
