package di

import (
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

type Injector struct {
	tag                     string
	ExtendTag               func(field reflect.StructField, value reflect.Value) (string, bool)
	OnGetByName             func(name string) (reflect.Value, bool)
	OnGetOneByInterfaceType func(p reflect.Type) reflect.Value
	OnGetByInterfaceType    func(p reflect.Type) []reflect.Value
}

func New(tag string) *Injector {
	return &Injector{
		tag: tag,
	}
}

func (i *Injector) Inject(c any) error {
	t := reflect.TypeOf(c)
	v := reflect.ValueOf(c)
	nodes := i.ScanNodes(t, v)
	if len(nodes) == 0 {
		return nil
	}
	return nil
}

func (i *Injector) ScanNodes(t reflect.Type, v reflect.Value) []*Node {
	var nodes []*Node
	_ = reflectx.ForEachFieldV2(t, v, true, func(field reflect.StructField, value reflect.Value) error {
		if tag, ok := field.Tag.Lookup(i.tag); ok {
			nodes = append(nodes, &Node{
				Tag:   tag,
				Type:  field.Type,
				Value: value,
			})
		} else if i.ExtendTag != nil {
			if tag, ok := i.ExtendTag(field, value); ok {
				nodes = append(nodes, &Node{
					Tag:   tag,
					Type:  field.Type,
					Value: value,
				})
			}
		} else if field.Anonymous && field.Type.Kind() == reflect.Struct {
			nodes = append(nodes, i.ScanNodes(field.Type, value)...)
		}
		return nil
	})
	return nodes
}
