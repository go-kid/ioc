package meta

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"reflect"
	"strings"
)

type Meta struct {
	Name         string
	Address      string
	Raw          interface{}
	Type         reflect.Type
	Value        reflect.Value
	Dependencies []*Dependency
	Properties   []*Property
	Produce      []*Meta
	DependsBy    []*Meta
}

func NewMeta(c interface{}) *Meta {
	if c == nil {
		panic("passed nil interface to ioc")
	}
	var (
		dependencies []*Dependency
		produces     []*Meta
		properties   []*Property
	)
	t := reflect.TypeOf(c)
	v := reflect.ValueOf(c)
	dependencies, produces, properties = analyseComponent(t, v)
	return &Meta{
		Name:         defination.GetComponentName(c),
		Address:      fmt.Sprintf("%p", c),
		Raw:          c,
		Type:         t,
		Value:        v,
		Dependencies: dependencies,
		Properties:   properties,
		Produce:      produces,
		DependsBy:    nil,
	}
}

func analyseComponent(t reflect.Type, v reflect.Value) (dependencies []*Dependency, produces []*Meta, properties []*Property) {
	_ = reflectx.ForEachFieldV2(t, v, true, func(field reflect.StructField, value reflect.Value) error {
		if prefix, ok := defination.IsConfigure(field, value); ok {
			properties = append(properties, &Property{
				Prefix: prefix,
				Type:   field.Type,
				Value:  value,
			})
		}
		if _, ok := defination.IsProduce(field); ok {
			v := reflectx.New(field.Type)
			reflectx.Set(value, v)
			p := NewMeta(value.Interface())
			produces = append(produces, p)
		}
		if name, ok := defination.IsDependency(field); ok {
			dependencies = append(dependencies, &Dependency{
				SpecifyName: name,
				Type:        field.Type,
				Value:       value,
			})
		} else if field.Anonymous && field.Type.Kind() == reflect.Struct {
			ds, ps, pps := analyseComponent(field.Type, value)
			dependencies = append(dependencies, ds...)
			produces = append(produces, ps...)
			properties = append(properties, pps...)
			return nil
		}
		return nil
	})
	return
}

func (m *Meta) ID() string {
	return fmt.Sprintf("%s(%s#%s)", m.Name, m.Type, m.Address)
}

func (m *Meta) DependBy(parent *Meta) {
	m.DependsBy = append(m.DependsBy, parent)
}

type kv struct {
	k string
	v string
}

func (m *Meta) DotNodeAttr() map[string]string {
	var label = []*kv{
		{k: "", v: m.Name},
		{k: "Type", v: m.Type.String()},
		{k: "Props", v: strings.Join(lo.Map[*Property, string](m.Properties, func(p *Property, _ int) string {
			return p.Prefix
		}), ", ")},
	}

	labels := lo.Map[*kv, string](label, func(item *kv, _ int) string {
		if item.k == "" {
			return item.v
		}
		return fmt.Sprintf("%s: %s", item.k, item.v)
	})
	return map[string]string{
		"label": StringEscape("{" + strings.Join(labels, "|") + "}"),
		"shape": "record",
	}
}

func StringEscape(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}
