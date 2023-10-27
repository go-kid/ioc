package meta

import (
	"fmt"
	"github.com/samber/lo"
	"reflect"
	"strings"
)

//
//const (
//	InjectTag  = "wire"
//	ProduceTag = "produce"
//	PropTag    = "prop"
//)

type Meta struct {
	Name         string
	Address      string
	Raw          interface{}
	Type         reflect.Type
	Value        reflect.Value
	Dependencies []*Node
	Properties   []*Node
	Produce      []*Meta
	DependsBy    []*Meta
}

//
//func NewMeta(c interface{}) *Meta {
//	if c == nil {
//		panic("passed nil interface to ioc")
//	}
//	t := reflect.TypeOf(c)
//	v := reflect.ValueOf(c)
//	return &Meta{
//		Name:         GetComponentName(v),
//		Address:      fmt.Sprintf("%p", c),
//		Raw:          c,
//		Type:         t,
//		Value:        v,
//		Dependencies: scanDependencies(t, v),
//		Properties:   scanProperties(t, v),
//		Produce:      scanProduces(t, v),
//		DependsBy:    nil,
//	}
//}
//
//func scanDependencies(t reflect.Type, v reflect.Value) []*Node {
//	wireInjector := scanner.New(InjectTag)
//	return wireInjector.ScanNodes(t, v)
//}
//
//func scanProduces(t reflect.Type, v reflect.Value) []*Meta {
//	productInjector := scanner.New(ProduceTag)
//	return lo.Map(productInjector.ScanNodes(t, v), func(item *Node, _ int) *Meta {
//		v := reflectx.New(item.Type)
//		reflectx.Set(item.Value, v)
//		p := NewMeta(item.Value.Interface())
//		return p
//	})
//}
//
//func scanProperties(t reflect.Type, v reflect.Value) []*Node {
//	propInjector := scanner.New(PropTag)
//	propInjector.ExtendTag = func(field reflect.StructField, value reflect.Value) (string, bool) {
//		if configuration, ok := value.Interface().(defination.Configuration); ok {
//			return configuration.Prefix(), true
//		}
//		return "", false
//	}
//	return propInjector.ScanNodes(t, v)
//}

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
		{k: "Props", v: strings.Join(lo.Map[*Node, string](m.Properties, func(p *Node, _ int) string {
			return p.Tag
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
