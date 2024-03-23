package configure

import (
	"encoding/json"
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/ioc/util/strconv2"
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

type executeExpressionPopulation struct {
	expReg *regexp.Regexp
	once   sync.Once
}

func (e *executeExpressionPopulation) Order() int {
	return OrderExecuteExpressionPopulation
}

func (e *executeExpressionPopulation) Filter(d *meta.Node) bool {
	e.once.Do(func() {
		e.expReg = regexp.MustCompile("\\$\\{\\w+(\\.\\w+)*(:[^{}]*)?}")
	})
	return e.expReg.MatchString(d.TagVal)
}

func (e *executeExpressionPopulation) Populate(r Binder, prop *meta.Node) error {
	rawTagVal := prop.TagVal
	prop.TagVal = e.expReg.ReplaceAllStringFunc(prop.TagVal, func(s string) string {
		exp := s[2 : len(s)-1]
		//split expression key and default value
		spExp := strings.SplitN(exp, ":", 2)
		exp = spExp[0]
		expVal := r.Get(exp)
		if expVal == nil {
			if len(spExp) != 2 {
				syslog.Fatalf("config path '%s' used by expression tag value is missing", exp)
			}
			//parse tag default value
			var err error
			expVal, err = strconv2.ParseAny(spExp[1])
			if err != nil {
				syslog.Fatalf("parse expression tag default value error: %v", spExp[1], err)
			}
		}
		val, err := marshalTagVal(expVal)
		if err != nil {
			syslog.Fatalf("marshal expression tag value %v error: %v", expVal, err)
		}
		return val
	})
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

func (p *propPopulation) Filter(d *meta.Node) bool {
	return d.Tag == defination.PropTag
}

func (p *propPopulation) Populate(r Binder, prop *meta.Node) error {
	return reflectx.SetValue(prop.Value, func(a any) error {
		return r.Unmarshall(prop.TagVal, a)
	})
}

type valuePopulation struct {
	once sync.Once
	hm   reflectx.Interceptor
}

func (v *valuePopulation) Order() int {
	return OrderValuePopulation
}

func (v *valuePopulation) Filter(d *meta.Node) bool {
	return d.Tag == defination.ValueTag
}

func (v *valuePopulation) Populate(r Binder, prop *meta.Node) error {
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
	})
	if prop.TagVal == "" {
		return nil
	}
	return reflectx.SetAnyValueFromString(prop.Type, prop.Value, prop.TagVal, v.hm)
}
