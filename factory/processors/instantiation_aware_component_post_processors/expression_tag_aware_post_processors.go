package instantiation_aware_component_post_processors

import (
	"fmt"
	"github.com/expr-lang/expr"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/el"
)

type expressionTagAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	definition.LazyInitComponent
	el el.Helper
}

func NewExpressionTagAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &expressionTagAwarePostProcessors{
		el: el.NewExpr(),
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
		if !c.el.MatchString(prop.TagVal) {
			continue
		}
		rawTagVal := prop.TagVal

		content, err := c.el.ReplaceAllContent(prop.TagVal, func(exp string) (string, error) {
			program, err := expr.Compile(exp)
			if err != nil {
				return "", fmt.Errorf("compile expression '%s' error: %v", exp, err)
			}
			result, err := expr.Run(program, nil)
			if err != nil {
				return "", fmt.Errorf("execute expression '%s' program error: %v", exp, err)
			}
			val, err := marshalTagVal(result)
			if err != nil {
				return "", fmt.Errorf("marshal expression tag value %v error: %v", result, err)
			}
			return val, nil
		})
		if err != nil {
			return nil, err
		}
		prop.TagVal = content
		syslog.Debugf("execute expression language '%s' -> '%s'", rawTagVal, prop.TagVal)
	}
	return nil, nil
}
