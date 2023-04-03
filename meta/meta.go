package meta

import (
	"fmt"
	"github.com/kidhat/kid-ioc/defination"
	"github.com/kidhat/kid-ioc/util/draw"
	"github.com/kidhat/kid-ioc/util/reflectx"
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
	Properties   []*property
	DependsBy    []*Meta
}

func NewMeta(c interface{}) *Meta {
	if c == nil {
		panic("passed nil interface to ioc")
	}
	var dependencies []*Dependency
	var properties []*property
	_ = reflectx.ForEachField(c, true, func(field reflect.StructField, value reflect.Value) error {
		if name, ok := defination.IsDependency(field); ok {
			dependencies = append(dependencies, &Dependency{
				SpecifyName: name,
				Type:        field.Type,
				Value:       value,
			})
		}
		if prefix, ok := defination.IsConfigure(field, value); ok {
			properties = append(properties, &property{
				Prefix: prefix,
				Type:   field.Type,
				Value:  value,
			})
		}
		return nil
	})
	return &Meta{
		Name:         defination.GetComponentName(c),
		Address:      fmt.Sprintf("%p", c),
		Raw:          c,
		Type:         reflect.TypeOf(c),
		Value:        reflect.ValueOf(c),
		Dependencies: dependencies,
		Properties:   properties,
	}
}

func (m *Meta) ID() string {
	return fmt.Sprintf("%s(%s#%s)", m.Name, m.Type, m.Address)
}

func (m *Meta) DependBy(parent *Meta) {
	m.DependsBy = append(m.DependsBy, parent)
}

func (m *Meta) String() string {
	var items = []string{
		"ID: " + m.ID(),
		"Address: " + m.Address,
		"Type: " + m.Type.String(),
		"Value: " + m.Value.String(),
		"Dependencies: " + func() string {
			arr := collect(m.Dependencies, func(o interface{}) string {
				return o.(*Dependency).Name()
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
				return o.(*Meta).ID()
			})
			return fmt.Sprintf("[%s]", strings.Join(arr, ", "))
		}(),
	}
	return draw.Frame(items)
}

func (m *Meta) Describe() {
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
