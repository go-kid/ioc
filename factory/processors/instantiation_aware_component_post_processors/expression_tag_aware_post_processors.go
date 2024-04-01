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

type expressionTagAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	Configure configure.Configure
	expReg    *regexp.Regexp
}

func NewExpressionTagAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &expressionTagAwarePostProcessors{
		expReg: regexp.MustCompile("\\$\\{\\w+(\\.\\w+)*(:[^{}]*)?}"),
	}
}

func (c *expressionTagAwarePostProcessors) PostProcessComponentFactory(factory factory.Factory) error {
	c.Configure = factory.GetConfigure()
	return nil
}

func (c *expressionTagAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (c *expressionTagAwarePostProcessors) Order() int {
	return 0
}

func (c *expressionTagAwarePostProcessors) PostProcessProperties(properties []*component_definition.Node, component any, componentName string) ([]*component_definition.Node, error) {
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
			if expVal == nil {
				if len(spExp) != 2 {
					return nil, fmt.Errorf("config path '%s' used by expression tag value is missing", exp)
				}
				//parse tag default value
				if defaultVal := spExp[1]; defaultVal == "" {
					prop.TagVal = strings.Replace(prop.TagVal, s, "", 1)
					continue
				} else {
					var err error
					expVal, err = strconv2.ParseAny(defaultVal)
					if err != nil {
						return nil, fmt.Errorf("parse expression tag default value %s error: %v", defaultVal, err)
					}
				}
			}
			val, err := marshalTagVal(expVal)
			if err != nil {
				return nil, fmt.Errorf("marshal expression tag value %v error: %v", expVal, err)
			}
			prop.TagVal = strings.Replace(prop.TagVal, s, val, 1)
		}
		syslog.Tracef("execute tag expression '%s' -> '%s'", rawTagVal, prop.TagVal)
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
