package processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/el"
	"github.com/go-kid/strconv2"
	"github.com/pkg/errors"
	"strings"
)

type configQuoteAwarePostProcessors struct {
	DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	definition.LazyInitComponent
	Configure configure.Configure
	el        el.Helper
}

func NewConfigQuoteAwarePostProcessors() container.InstantiationAwareComponentPostProcessor {
	return &configQuoteAwarePostProcessors{
		el: el.NewQuote(),
	}
}

func (c *configQuoteAwarePostProcessors) PostProcessComponentFactory(factory container.Factory) error {
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
	logger := syslog.Pref("ConfigQuoteAwarePostProcessor")
	for _, prop := range properties {
		if !c.el.MatchString(prop.TagStr) {
			continue
		}

		content, err := c.el.ReplaceAllContent(prop.TagStr, func(exp string) (string, error) {
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
				var defaultValue string
				if len(spExp) == 2 {
					defaultValue = spExp[1]
				} else {
					logger.Warnf("config quote value '%s' is neither in configuration nor has a default value", exp)
				}
				//parse tag default value
				if defaultValue != "" {
					parsedVal, err := strconv2.ParseAny(defaultValue)
					if err != nil {
						return "", errors.Wrapf(err, "parse config quote default value '%s' error", defaultValue)
					}
					expVal = parsedVal
				}
			}
			prop.SetConfiguration(exp, expVal)

			if expVal == nil {
				return "", nil
			}
			marshalVal, err := strconv2.FormatAny(expVal)
			if err != nil {
				return "", errors.Wrapf(err, "marshal expression tag value %v error", expVal)
			}

			return marshalVal, nil
		})
		if err != nil {
			return nil, errors.WithMessagef(err, "config quote value on '%s' failed", prop)
		}

		prop.TagVal = content
		logger.Debugf("config quote value on '%s'\n '%s' -> '%s'", prop, prop.TagStr, prop.TagVal)
	}
	return nil, nil
}
