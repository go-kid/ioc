package configure

import (
	"encoding/json"
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/ioc/util/strconv2"
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

const (
	MinBuiltinPopulationOrder = iota + 1000
	OrderExecuteExpressionPopulation
	OrderPropPopulation
	OrderValuePopulation
)

const (
	ArgValidate component_definition.ArgType = "validate"
)

type executeExpressionPopulation struct {
	expReg *regexp.Regexp
	once   sync.Once
}

func (e *executeExpressionPopulation) Order() int {
	return OrderExecuteExpressionPopulation
}

func (e *executeExpressionPopulation) Filter(d *component_definition.Node) bool {
	e.once.Do(func() {
		e.expReg = regexp.MustCompile("\\$\\{\\w+(\\.\\w+)*(:[^{}]*)?}")
	})
	return e.expReg.MatchString(d.TagVal)
}

func (e *executeExpressionPopulation) Populate(r Binder, prop *component_definition.Node) error {
	rawTagVal := prop.TagVal

	matches := e.expReg.FindAllString(prop.TagVal, -1)
	for _, s := range matches {
		exp := s[2 : len(s)-1]
		//split expression key and default value
		spExp := strings.SplitN(exp, ":", 2)
		exp = spExp[0]
		expVal := r.Get(exp)
		if expVal == nil {
			if len(spExp) != 2 {
				return fmt.Errorf("config path '%s' used by expression tag value is missing", exp)
			}
			//parse tag default value
			if defaultVal := spExp[1]; defaultVal == "" {
				prop.TagVal = strings.Replace(prop.TagVal, s, "", 1)
				continue
			} else {
				var err error
				expVal, err = strconv2.ParseAny(defaultVal)
				if err != nil {
					return fmt.Errorf("parse expression tag default value %s error: %v", defaultVal, err)
				}
			}
		}
		val, err := marshalTagVal(expVal)
		if err != nil {
			return fmt.Errorf("marshal expression tag value %v error: %v", expVal, err)
		}
		prop.TagVal = strings.Replace(prop.TagVal, s, val, 1)
	}
	syslog.Tracef("execute tag expression '%s' -> '%s'", rawTagVal, prop.TagVal)
	return nil
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

type propPopulation struct {
}

func (p *propPopulation) Order() int {
	return OrderPropPopulation
}

func (p *propPopulation) Filter(d *component_definition.Node) bool {
	return d.Tag == definition.PropTag
}

func (p *propPopulation) Populate(r Binder, prop *component_definition.Node) error {
	if r.Get(prop.TagVal) == nil {
		if prop.Args().Has(component_definition.ArgRequired, "true") {
			return fmt.Errorf("properties is required")
		}
		return nil
	}
	err := reflectx.SetValue(prop.Value, func(a any) error {
		return r.Unmarshall(prop.TagVal, a)
	})
	if err != nil {
		return fmt.Errorf("population 'prop' value %s to %s failed: %v", prop.TagVal, prop.Type.String(), err)
	}
	return validate(prop)
}

type valuePopulation struct {
	once     sync.Once
	validate *validator.Validate
	hm       reflectx.Interceptor
}

func (v *valuePopulation) Order() int {
	return OrderValuePopulation
}

func (v *valuePopulation) Filter(d *component_definition.Node) bool {
	return d.Tag == definition.ValueTag
}

func (v *valuePopulation) Populate(r Binder, prop *component_definition.Node) error {
	v.once.Do(func() {
		var jsonUnmarshalHandler = func(_ reflect.Type, v reflect.Value, s string) error {
			return reflectx.SetValue(v, func(a any) error {
				return json.Unmarshal([]byte(s), a)
			})
		}
		v.hm = reflectx.Interceptor{
			reflect.Map:       jsonUnmarshalHandler,
			reflect.Slice:     jsonUnmarshalHandler,
			reflect.Struct:    jsonUnmarshalHandler,
			reflect.Interface: jsonUnmarshalHandler,
		}
		v.validate = validator.New(validator.WithRequiredStructEnabled())
	})
	if prop.TagVal == "" {
		if prop.Args().Has(component_definition.ArgRequired, "true") {
			return fmt.Errorf("properties is required")
		}
		return nil
	}
	err := reflectx.SetAnyValueFromString(prop.Type, prop.Value, prop.TagVal, v.hm)
	if err != nil {
		return fmt.Errorf("population 'value' value %s to %s failed: %v", prop.TagVal, prop.Type.String(), err)
	}
	return validate(prop)
}

var v = validator.New(validator.WithRequiredStructEnabled())

func validate(prop *component_definition.Node) error {
	if ts, ok := prop.Args().Find(ArgValidate); ok {
		var p = prop.Type
		if p.Kind() == reflect.Pointer {
			p = p.Elem()
		}
		if p.Kind() == reflect.Struct {
			return v.Struct(prop.Value.Interface())
		} else {
			for _, t := range ts {
				err := v.Var(prop.Value.Interface(), t)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
