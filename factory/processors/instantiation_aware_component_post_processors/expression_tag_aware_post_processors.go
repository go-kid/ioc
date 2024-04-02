package instantiation_aware_component_post_processors

import (
	"fmt"
	"github.com/expr-lang/expr"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/syslog"
	"regexp"
	"strings"
)

type expressionTagAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	expReg *regexp.Regexp
}

func NewExpressionTagAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &expressionTagAwarePostProcessors{
		expReg: regexp.MustCompile("#{[^{}]*}"),
	}
}

func (c *expressionTagAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (c *expressionTagAwarePostProcessors) Order() int {
	return PriorityOrderPropertyExpressionTagAware
}

func (c *expressionTagAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if !c.expReg.MatchString(prop.TagVal) {
			continue
		}
		rawTagVal := prop.TagVal

		matches := c.expReg.FindAllString(prop.TagVal, -1)
		for _, s := range matches {
			exp := s[2 : len(s)-1]
			program, err := expr.Compile(exp)
			if err != nil {
				return nil, fmt.Errorf("compile expression '%s' error: %v", exp, err)
			}
			result, err := expr.Run(program, nil)
			if err != nil {
				return nil, fmt.Errorf("execute expression '%s' program error: %v", exp, err)
			}
			val, err := marshalTagVal(result)
			if err != nil {
				return nil, fmt.Errorf("marshal expression tag value %v error: %v", result, err)
			}
			prop.TagVal = strings.Replace(prop.TagVal, s, val, 1)
		}
		syslog.Debugf("execute tag expression '%s' -> '%s'", rawTagVal, prop.TagVal)
	}
	return nil, nil
}
