package configure

import (
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
		spExp := strings.SplitN(exp, ":", 2)
		exp = spExp[0]
		expVal := r.Get(exp)
		if expVal == nil {
			if len(spExp) == 2 {
				return spExp[1]
			}
			syslog.Fatalf("config path '%s' used by expression tag value is missing", exp)
		}
		switch expVal.(type) {
		case string:
			return expVal.(string)
		default:
			return fmt.Sprintf("%v", expVal)
		}
	})
	syslog.Tracef("execute tag expression '%s' -> '%s'", rawTagVal, prop.TagVal)
	return nil
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
	//syslog.Tracef("viper binder start bind config %s, prefix: %s", prop.ID(), prop.TagVal)
	var fieldType = prop.Type
	var isPtrType = false
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
		isPtrType = true
	}
	var val = reflect.New(fieldType)
	err := r.Unmarshall(prop.TagVal, val.Interface())
	if err != nil {
		return fmt.Errorf("populate prop config %s.Value(%s) error: %v", prop.ID(), prop.TagVal, err)
	}
	if isPtrType {
		prop.Value.Set(val)
	} else {
		prop.Value.Set(val.Elem())
	}

	return nil
}

type valuePopulation struct {
}

func (v *valuePopulation) Order() int {
	return OrderValuePopulation
}

func (v *valuePopulation) Filter(d *meta.Node) bool {
	return d.Tag == defination.ValueTag
}

func (v *valuePopulation) Populate(r Binder, d *meta.Node) error {
	parseAny, err := strconv2.ParseAny(d.TagVal)
	if err != nil {
		return err
	}
	fmt.Println(parseAny)
	return reflectx.SetAnyValue(d.Type, d.Value, d.TagVal)
}
