package scanner

import (
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

type Scanner struct {
	tag       string
	ExtendTag func(field reflect.StructField, value reflect.Value) (string, bool)
}

func New(tag string) *Scanner {
	return &Scanner{
		tag: tag,
	}
}

func (i *Scanner) ScanNodes(t reflect.Type, v reflect.Value) []*injector.Node {
	var nodes []*injector.Node
	_ = reflectx.ForEachFieldV2(t, v, true, func(field reflect.StructField, value reflect.Value) error {
		if tag, ok := field.Tag.Lookup(i.tag); ok {
			nodes = append(nodes, &injector.Node{
				Tag:   tag,
				Type:  field.Type,
				Value: value,
			})
		} else if i.ExtendTag != nil {
			if tag, ok := i.ExtendTag(field, value); ok {
				nodes = append(nodes, &injector.Node{
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
