package configure

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

const (
	OrderPropPopulation = iota + 1000
	OrderValuePopulation
)

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
	return reflectx.SetAnyValue(d.Type, d.Value, d.TagVal)
}
