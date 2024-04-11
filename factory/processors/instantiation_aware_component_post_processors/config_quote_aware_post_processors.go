package instantiation_aware_component_post_processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/el"
	"github.com/go-kid/ioc/util/strconv2"
	"strings"
)

type configQuoteAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	definition.LazyInitComponent
	Configure configure.Configure
	el        el.Helper
}

func NewConfigQuoteAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &configQuoteAwarePostProcessors{
		el: el.NewQuote(),
	}
}

func (c *configQuoteAwarePostProcessors) PostProcessComponentFactory(factory factory.Factory) error {
	c.Configure = factory.GetConfigure()
	return nil
}

func (c *configQuoteAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (c *configQuoteAwarePostProcessors) Order() int {
	return PriorityOrderPropertyConfigQuoteAware
}

func (c *configQuoteAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if !c.el.MatchString(prop.TagVal) {
			continue
		}
		rawTagVal := prop.TagVal

		content, err := c.el.ReplaceAllContent(prop.TagVal, func(exp string) (string, error) {
			//split expression key and default value
			spExp := strings.SplitN(exp, ":", 2)
			exp = spExp[0]
			expVal := c.Configure.Get(exp)
			useDefaultValue := false
			if expVal == nil {
				useDefaultValue = true
			} else if m, ok := expVal.(map[string]any); ok && len(m) == 0 {
				useDefaultValue = true
			} else if arr, ok := expVal.([]any); ok && len(arr) == 0 {
				useDefaultValue = true
			}
			if useDefaultValue {
				if len(spExp) != 2 {
					return "", fmt.Errorf("config quote value '%s' is not exist and has no default value", exp)
				}
				//parse tag default value
				if defaultVal := spExp[1]; defaultVal == "" {
					return "", nil
				} else {
					var err error
					expVal, err = strconv2.ParseAny(defaultVal)
					if err != nil {
						return "", fmt.Errorf("parse config quote default value %s error: %v", defaultVal, err)
					}
				}
			}

			marshalVal, err := strconv2.FormatAny(expVal)
			if err != nil {
				return "", fmt.Errorf("marshal expression tag value %v error: %v", expVal, err)
			}

			_ = prop.SetConfiguration(exp, expVal, false)

			return marshalVal, nil
		})
		if err != nil {
			return nil, fmt.Errorf("config quote value on '%s' failed: %v", prop.ID(), err)
		}

		prop.TagVal = content
		syslog.Pref("ConfigQuoteAwarePostProcessor").Debugf("config quote value on '%s'\n '%s' -> '%s'", prop.ID(), rawTagVal, prop.TagVal)
	}
	return nil, nil
}
