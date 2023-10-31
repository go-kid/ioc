package scanner

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"reflect"
)

type scanner struct {
	tags []string
}

func New(tags ...string) Scanner {
	return &scanner{tags: tags}
}

func (s *scanner) ScanComponent(c any) *meta.Meta {
	if c == nil {
		panic("passed nil interface to ioc")
	}
	return s.newMeta(c)
}

func (s *scanner) newMeta(c any) *meta.Meta {
	t := reflect.TypeOf(c)
	v := reflect.ValueOf(c)
	return &meta.Meta{
		Name:            meta.GetComponentName(v),
		Address:         fmt.Sprintf("%p", c),
		Raw:             c,
		Type:            t,
		Value:           v,
		Dependencies:    s.scanDependencies(t, v),
		Properties:      s.scanProperties(t, v),
		Produce:         s.scanProduces(t, v),
		DependsBy:       nil,
		CustomizedField: s.scanCustomizedField(t, v),
	}
}

func (s *scanner) scanDependencies(t reflect.Type, v reflect.Value) []*meta.Node {
	return s.ScanNodes(meta.InjectTag, t, v)
}

func (s *scanner) scanProduces(t reflect.Type, v reflect.Value) []*meta.Meta {
	return lo.Map(s.ScanNodes(meta.ProduceTag, t, v), func(item *meta.Node, _ int) *meta.Meta {
		v := reflectx.New(item.Type)
		reflectx.Set(item.Value, v)
		p := s.newMeta(item.Value.Interface())
		return p
	})
}

func (s *scanner) scanProperties(t reflect.Type, v reflect.Value) []*meta.Node {
	var tag = "Configuration.Prefix"
	configureHandler := func(field reflect.StructField, value reflect.Value) (string, string, bool) {
		if configuration, ok := value.Interface().(defination.Configuration); ok {
			return tag, configuration.Prefix(), true
		}
		return "", "", false
	}
	return s.ScanNodes(meta.PropTag, t, v, configureHandler)
}

func (s *scanner) scanCustomizedField(t reflect.Type, v reflect.Value) []*meta.Node {
	return lo.FlatMap(s.tags, func(tag string, _ int) []*meta.Node {
		return s.ScanNodes(tag, t, v)
	})
}

type ExtTagHandler func(field reflect.StructField, value reflect.Value) (string, string, bool)

func (s *scanner) ScanNodes(tag string, t reflect.Type, v reflect.Value, handlers ...ExtTagHandler) []*meta.Node {
	var nodes []*meta.Node
	_ = reflectx.ForEachFieldV2(t, v, true, func(field reflect.StructField, value reflect.Value) error {
		//find tag in struct field tag
		if tag != "" {
			if tagVal, ok := field.Tag.Lookup(tag); ok {
				nodes = append(nodes, &meta.Node{
					Field:   field,
					Tag:     tag,
					TagVal:  tagVal,
					Type:    field.Type,
					Value:   value,
					Injects: nil,
				})
				return nil
			}
		}
		//if is embed struct, find inside
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			nodes = append(nodes, s.ScanNodes(tag, field.Type, value, handlers...)...)
			return nil
		}
		//use first success extra tag handler
		if len(handlers) > 0 {
			for _, handler := range handlers {
				if tag, tagVal, ok := handler(field, value); ok {
					nodes = append(nodes, &meta.Node{
						Field:   field,
						Tag:     tag,
						TagVal:  tagVal,
						Type:    field.Type,
						Value:   value,
						Injects: nil,
					})
					return nil
				}
			}
		}
		return nil
	})
	return nodes
}
