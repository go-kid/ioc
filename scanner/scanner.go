package scanner

import (
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"reflect"
)

type scanner struct {
	policies []ScanPolicy
}

func Default() Scanner {
	return &scanner{}
}

func (s *scanner) AddTags(policies []ScanPolicy) {
	s.policies = append(s.policies, policies...)
}

func (s *scanner) ScanComponent(c any) *meta.Meta {
	m := meta.NewMeta(c)
	m.Dependencies = s.scanDependencies(m)
	m.Properties = s.scanProperties(m)
	m.CustomizedField = s.scanCustomizedField(m)
	return m
}

func (s *scanner) scanDependencies(m *meta.Meta) []*meta.Node {
	return s.ScanNodes(meta.NewHolder(m), DefaultScanPolicy(defination.InjectTag, nil))
}

func (s *scanner) scanProperties(m *meta.Meta) []*meta.Node {
	return s.ScanNodes(meta.NewHolder(m), DefaultScanPolicy(defination.PropTag, configureHandler))
}

func configureHandler(_ reflect.StructField, value reflect.Value) (string, string, bool) {
	var tag = "Configuration.Prefix"
	if configuration, ok := value.Interface().(defination.Configuration); ok {
		return tag, configuration.Prefix(), true
	}
	return "", "", false
}

func (s *scanner) scanCustomizedField(m *meta.Meta) []*meta.Node {
	return lo.FlatMap(s.policies, func(sp ScanPolicy, _ int) []*meta.Node {
		return s.ScanNodes(meta.NewHolder(m), sp)
	})
}

func (s *scanner) ScanNodes(holder *meta.Holder, sp ScanPolicy) []*meta.Node {
	var nodes []*meta.Node
	_ = reflectx.ForEachFieldV2(holder.Type, holder.Value, false, func(field reflect.StructField, value reflect.Value) error {
		var base = &meta.Base{
			Type:  field.Type,
			Value: value,
		}
		//if is embed struct, find inside
		if field.Anonymous && field.Tag == "" && field.Type.Kind() == reflect.Struct {
			var embedSource = meta.NewEmbedHolder(base, holder)
			nodes = append(nodes, s.ScanNodes(embedSource, sp)...)
			return nil
		}

		if !value.CanSet() {
			return nil
		}

		//find tag in struct field tag
		if tag := sp.Tag(); tag != "" {
			if tagVal, ok := field.Tag.Lookup(tag); ok {
				nodes = append(nodes, meta.NewNode(base, holder, field, tag, tagVal))
				return nil
			}
		}

		//if not find in tag, use extract tag handler
		if handler := sp.ExtHandler(); handler != nil {
			if tag, tagVal, ok := handler(field, value); ok {
				nodes = append(nodes, meta.NewNode(base, holder, field, tag, tagVal))
				return nil
			}
		}
		return nil
	})
	return nodes
}
