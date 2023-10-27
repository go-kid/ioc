package scanner

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"reflect"
)

type Scanner struct {
}

func New() *Scanner {
	return &Scanner{}
}

func (s *Scanner) ScanComponent(c any) *meta.Meta {
	if c == nil {
		panic("passed nil interface to ioc")
	}
	return s.newMeta(c)
}

func (s *Scanner) newMeta(c any) *meta.Meta {
	t := reflect.TypeOf(c)
	v := reflect.ValueOf(c)
	return &meta.Meta{
		Name:         meta.GetComponentName(v),
		Address:      fmt.Sprintf("%p", c),
		Raw:          c,
		Type:         t,
		Value:        v,
		Dependencies: s.scanDependencies(t, v),
		Properties:   s.scanProperties(t, v),
		Produce:      s.scanProduces(t, v),
		DependsBy:    nil,
	}
}

const (
	InjectTag  = "wire"
	ProduceTag = "produce"
	PropTag    = "prop"
)

func (s *Scanner) scanDependencies(t reflect.Type, v reflect.Value) []*meta.Node {
	return s.ScanNodes(InjectTag, t, v)
}

func (s *Scanner) scanProduces(t reflect.Type, v reflect.Value) []*meta.Meta {
	return lo.Map(s.ScanNodes(ProduceTag, t, v), func(item *meta.Node, _ int) *meta.Meta {
		v := reflectx.New(item.Type)
		reflectx.Set(item.Value, v)
		p := s.newMeta(item.Value.Interface())
		return p
	})
}

func (s *Scanner) scanProperties(t reflect.Type, v reflect.Value) []*meta.Node {
	configureHandler := func(field reflect.StructField, value reflect.Value) (string, bool) {
		if configuration, ok := value.Interface().(defination.Configuration); ok {
			return configuration.Prefix(), true
		}
		return "", false
	}
	return s.ScanNodes(PropTag, t, v, configureHandler)
}

type ExtTagHandler func(field reflect.StructField, value reflect.Value) (string, bool)

func (s *Scanner) ScanNodes(tag string, t reflect.Type, v reflect.Value, handlers ...ExtTagHandler) []*meta.Node {
	var nodes []*meta.Node
	_ = reflectx.ForEachFieldV2(t, v, true, func(field reflect.StructField, value reflect.Value) error {
		//find tag in struct field tag
		if tagVal, ok := field.Tag.Lookup(tag); ok {
			nodes = append(nodes, &meta.Node{
				Tag:   tagVal,
				Type:  field.Type,
				Value: value,
			})
			return nil
		}
		//if is embed struct, find inside
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			nodes = append(nodes, s.ScanNodes(tag, field.Type, value, handlers...)...)
			return nil
		}
		//use first success extra tag handler
		if len(handlers) > 0 {
			for _, handler := range handlers {
				if tagVal, ok := handler(field, value); ok {
					nodes = append(nodes, &meta.Node{
						Tag:   tagVal,
						Type:  field.Type,
						Value: value,
					})
					return nil
				}
			}
		}
		return nil
	})
	return nodes
}
