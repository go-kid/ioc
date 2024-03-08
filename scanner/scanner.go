package scanner

import (
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
	m := meta.NewMeta(c)
	m.Dependencies = s.scanDependencies(m)
	m.Properties = s.scanProperties(m)
	m.CustomizedField = s.scanCustomizedField(m)
	return m
}

func (s *scanner) scanDependencies(source *meta.Meta) []*meta.Node {
	return s.ScanNodes(&meta.Source{
		Base: source.Base,
		Meta: source,
	}, meta.InjectTag)
}

func (s *scanner) scanProperties(source *meta.Meta) []*meta.Node {
	var tag = "Configuration.Prefix"
	configureHandler := func(field reflect.StructField, value reflect.Value) (string, string, bool) {
		if configuration, ok := value.Interface().(defination.Configuration); ok {
			return tag, configuration.Prefix(), true
		}
		return "", "", false
	}
	return s.ScanNodes(&meta.Source{
		Base: source.Base,
		Meta: source,
	}, meta.PropTag, configureHandler)
}

func (s *scanner) scanCustomizedField(source *meta.Meta) []*meta.Node {
	return lo.FlatMap(s.tags, func(tag string, _ int) []*meta.Node {
		return s.ScanNodes(&meta.Source{
			Base: source.Base,
			Meta: source,
		}, tag)
	})
}

type ExtTagHandler func(field reflect.StructField, value reflect.Value) (tag string, tagVal string, err bool)

func (s *scanner) ScanNodes(source *meta.Source, tag string, handlers ...ExtTagHandler) []*meta.Node {
	var nodes []*meta.Node
	_ = reflectx.ForEachFieldV2(source.Type, source.Value, false, func(field reflect.StructField, value reflect.Value) error {
		var base = &meta.Base{
			Type:  field.Type,
			Value: value,
		}
		//if is embed struct, find inside
		if field.Anonymous && field.Tag == "" && field.Type.Kind() == reflect.Struct {
			var embedSource = &meta.Source{
				Base:    base,
				Meta:    source.Meta,
				IsEmbed: true,
				Source:  source,
			}
			nodes = append(nodes, s.ScanNodes(embedSource, tag, handlers...)...)
			return nil
		}

		if !value.CanSet() {
			return nil
		}

		var newNodeFn = func(tag, tagVal string) *meta.Node {
			return &meta.Node{
				Base:   base,
				Source: source,
				Field:  field,
				Tag:    tag,
				TagVal: tagVal,
			}
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
