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

func Default() Scanner {
	return &scanner{}
}

func (s *scanner) AddTags(tags []string) {
	s.tags = append(s.tags, tags...)
}

func (s *scanner) ScanComponent(c any) *meta.Meta {
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
		DependsBy:       nil,
		CustomizedField: s.scanCustomizedField(t, v),
	}
}

func (s *scanner) scanDependencies(t reflect.Type, v reflect.Value) []*meta.Node {
	return s.ScanNodes(&meta.Source{
		Type:  t,
		Value: v,
	}, meta.InjectTag)
}

func (s *scanner) scanProperties(t reflect.Type, v reflect.Value) []*meta.Node {
	var tag = "Configuration.Prefix"
	configureHandler := func(field reflect.StructField, value reflect.Value) (string, string, bool) {
		if configuration, ok := value.Interface().(defination.Configuration); ok {
			return tag, configuration.Prefix(), true
		}
		return "", "", false
	}
	return s.ScanNodes(&meta.Source{
		Type:  t,
		Value: v,
	}, meta.PropTag, configureHandler)
}

func (s *scanner) scanCustomizedField(t reflect.Type, v reflect.Value) []*meta.Node {
	return lo.FlatMap(s.tags, func(tag string, _ int) []*meta.Node {
		return s.ScanNodes(&meta.Source{
			Type:  t,
			Value: v,
		}, tag)
	})
}

type ExtTagHandler func(field reflect.StructField, value reflect.Value) (string, string, bool)

func (s *scanner) ScanNodes(source *meta.Source, tag string, handlers ...ExtTagHandler) []*meta.Node {
	var nodes []*meta.Node
	_ = reflectx.ForEachFieldV2(source.Type, source.Value, false, func(field reflect.StructField, value reflect.Value) error {
		var newNodeFn = func(tag, tagVal string) *meta.Node {
			return &meta.Node{
				Source:  source,
				Field:   field,
				Tag:     tag,
				TagVal:  tagVal,
				Type:    field.Type,
				Value:   value,
				Injects: nil,
			}
		}
		//if is embed struct, find inside
		if field.Anonymous && field.Tag == "" && field.Type.Kind() == reflect.Struct {
			var child = &meta.Source{
				Source:      source,
				IsAnonymous: true,
				Type:        field.Type,
				Value:       value,
			}
			nodes = append(nodes, s.ScanNodes(child, tag, handlers...)...)
			return nil
		}

		if !value.CanSet() {
			return nil
		}

		//find tag in struct field tag
		if tagVal, ok := field.Tag.Lookup(tag); ok {
			nodes = append(nodes, newNodeFn(tag, tagVal))
			return nil
		}

		//use first success extra tag handler
		if len(handlers) > 0 {
			for _, handler := range handlers {
				if tag, tagVal, ok := handler(field, value); ok {
					nodes = append(nodes, newNodeFn(tag, tagVal))
					return nil
				}
			}
		}
		return nil
	})
	return nodes
}
