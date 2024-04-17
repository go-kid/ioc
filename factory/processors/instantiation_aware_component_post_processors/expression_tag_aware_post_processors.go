package instantiation_aware_component_post_processors

import (
	"github.com/expr-lang/expr"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/el"
	"github.com/go-kid/strconv2"
	"github.com/pkg/errors"
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
				return "", errors.Wrapf(err, "compile expression '%s' error", exp)
			}
			result, err := expr.Run(program, nil)
			if err != nil {
				return "", errors.Wrapf(err, "execute expression '%s' program error", exp)
			}
			val, err := strconv2.FormatAny(result)
			if err != nil {
				return "", errors.Wrapf(err, "marshal expression tag value %v error", result)
			}
			return val, nil
		})
		if err != nil {
			return nil, errors.WithMessagef(err, "execute expression language on '%s' failed", prop)
		}
		prop.TagVal = content
		syslog.Pref("ExpressionTagAwarePostProcessor").Debugf("execute expression language on '%s'\n '%s' -> '%s'", prop, rawTagVal, prop.TagVal)
	}
	return nil, nil
}
