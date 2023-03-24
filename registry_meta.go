package ioc

import (
	"fmt"
	"github.com/kid-hash/kid-ioc/util/draw"
	"github.com/kid-hash/kid-ioc/util/reflectx"
	"reflect"
	"strings"
)

type dependency struct {
	SpecifyName string
	Type        reflect.Type
	Value       reflect.Value
}

func (d *dependency) Name() string {
	if d.SpecifyName != "" {
		return d.SpecifyName
	}
	if v := d.Value.Interface(); v != nil {
		return getComponentName(v)
	}
	return getComponentName(reflect.New(d.Type).Interface())
}

type property struct {
	Prefix string
	Type   reflect.Type
	Value  reflect.Value
}

type meta struct {
	Name         string
	Address      string
	Raw          interface{}
	Type         reflect.Type
	Value        reflect.Value
	Dependencies []*dependency
	Properties   []*property
	DependsBy    []*meta
}

func newMeta(c interface{}) *meta {
	if c == nil {
		panic("passed nil interface to ioc")
	}
	var dependencies []*dependency
	var properties []*property
	_ = reflectx.ForEachField(c, true, func(field reflect.StructField, value reflect.Value) error {
		if name, ok := isDependency(field); ok {
			dependencies = append(dependencies, &dependency{
				SpecifyName: name,
				Type:        field.Type,
				Value:       value,
			})
		}
		if prefix, ok := isConfigure(field, value); ok {
			properties = append(properties, &property{
				Prefix: prefix,
				Type:   field.Type,
				Value:  value,
			})
		}
		return nil
	})
	return &meta{
		Name:         getComponentName(c),
		Address:      fmt.Sprintf("%p", c),
		Raw:          c,
		Type:         reflect.TypeOf(c),
		Value:        reflect.ValueOf(c),
		Dependencies: dependencies,
		Properties:   properties,
	}
}

func (m *meta) ID() string {
	return fmt.Sprintf("%s(%s#%s)", m.Name, m.Type, m.Address)
}

func (m *meta) DependBy(parent *meta) {
	m.DependsBy = append(m.DependsBy, parent)
}

func (m *meta) String() string {
	var items = []string{
		"ID: " + m.ID(),
		"Address: " + m.Address,
		"Type: " + m.Type.String(),
		"Value: " + m.Value.String(),
		"Dependencies: " + func() string {
			arr := collect(m.Dependencies, func(o interface{}) string {
				return o.(*dependency).Name()
			})
			return fmt.Sprintf("[%s]", strings.Join(arr, ", "))
		}(),
		"Properties: " + func() string {
			arr := collect(m.Properties, func(o interface{}) string {
				return o.(*property).Prefix
			})
			return fmt.Sprintf("[%s]", strings.Join(arr, ", "))
		}(),
		"DependsBy: " + func() string {
			arr := collect(m.DependsBy, func(o interface{}) string {
				return o.(*meta).ID()
			})
			return fmt.Sprintf("[%s]", strings.Join(arr, ", "))
		}(),
	}
	return draw.Frame(items)
}

func (m *meta) Describe() {
	fmt.Println(m.String())
}

func collect(arr interface{}, f func(o interface{}) string) []string {
	var sl []string
	val := reflect.ValueOf(arr)
	for i := 0; i < val.Len(); i++ {
		sl = append(sl, f(val.Index(i).Interface()))
	}
	return sl
}
