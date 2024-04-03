package instantiation_aware_component_post_processors

import (
	"encoding/json"
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/strconv2"
	"regexp"
	"strings"
)

type configQuoteAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	definition.LazyInitComponent
	Configure configure.Configure
	expReg    *regexp.Regexp
}

func NewConfigQuoteAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &configQuoteAwarePostProcessors{
		expReg: regexp.MustCompile("\\${[^{}]*}"),
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
		if !c.expReg.MatchString(prop.TagVal) {
			continue
		}
		rawTagVal := prop.TagVal

		matches := c.expReg.FindAllString(prop.TagVal, -1)
		for _, s := range matches {
			exp := s[2 : len(s)-1]
			//split expression key and default value
			spExp := strings.SplitN(exp, ":", 2)
			exp = spExp[0]
			expVal := c.Configure.Get(exp)
			useDefaultValue := false
			if expVal == nil {
				useDefaultValue = true
			} else if m, ok := expVal.(map[string]any); ok && len(m) == 0 {
				useDefaultValue = true
			}
			if useDefaultValue {
				if len(spExp) != 2 {
					return nil, fmt.Errorf("config quote value '%s' is not exist and has no default value", exp)
				}
				//parse tag default value
				if defaultVal := spExp[1]; defaultVal == "" {
					prop.TagVal = strings.Replace(prop.TagVal, s, "", 1)
					continue
				} else {
					var err error
					expVal, err = strconv2.ParseAny(defaultVal)
					if err != nil {
						return nil, fmt.Errorf("parse config quote default value %s error: %v", defaultVal, err)
					}
				}
			}
			val, err := marshalTagVal(expVal)
			if err != nil {
				return nil, fmt.Errorf("marshal expression tag value %v error: %v", expVal, err)
			}
			prop.TagVal = strings.Replace(prop.TagVal, s, val, 1)
		}
		syslog.Debugf("config quote value '%s' -> '%s'", rawTagVal, prop.TagVal)
	}
	return nil, nil
}

func marshalTagVal(expVal any) (string, error) {
	switch expVal.(type) {
	case string:
		return expVal.(string), nil
	case map[string]any, []any:
		bytes, err := json.Marshal(expVal)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	default:
		return fmt.Sprintf("%v", expVal), nil
	}
}
