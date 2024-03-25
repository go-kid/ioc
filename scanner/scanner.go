package scanner

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

type scanner struct {
	policies []ScanPolicy
}

func NewScanner(policies ...ScanPolicy) Scanner {
	s := &scanner{}
	s.AddScanPolicies(policies...)
	return s
}

func Default() Scanner {
	return NewScanner(
		//configuration policy
		NewConfigurationScanPolicy(definition.PropTag, propHandler),
		NewConfigurationScanPolicy(definition.ValueTag, nil),
		//component policy
		NewComponentScanPolicy(definition.InjectTag, nil),
	)
}

func (s *scanner) AddScanPolicies(policies ...ScanPolicy) {
	s.policies = append(s.policies, policies...)
}

func (s *scanner) ScanComponent(c any) *component_definition.Meta {
	m := component_definition.NewMeta(c)
	s.scanNodes(component_definition.NewHolder(m))
	return m
}

func (s *scanner) scanNodes(holder *component_definition.Holder) {
	_ = reflectx.ForEachFieldV2(holder.Type, holder.Value, false, func(field reflect.StructField, value reflect.Value) error {
		var base = &component_definition.Base{
			Type:  field.Type,
			Value: value,
		}
		//if is embed struct, find inside
		if field.Anonymous && field.Tag == "" && field.Type.Kind() == reflect.Struct {
			s.scanNodes(component_definition.NewEmbedHolder(base, holder))
			return nil
		}

		if !value.CanSet() {
			return nil
		}

		for _, sp := range s.policies {
			nt := sp.Group()
			//find tag in struct field tag
			if tag := sp.Tag(); tag != "" {
				if tagVal, ok := field.Tag.Lookup(tag); ok {
					holder.Meta.SetNodes(nt, component_definition.NewNode(base, holder, field, tag, tagVal))
					continue
				}
			}

			//if not find in tag, use extract tag handler
			if handler := sp.ExtHandler(); handler != nil {
				if tag, tagVal, ok := handler(field, value); ok {
					holder.Meta.SetNodes(nt, component_definition.NewNode(base, holder, field, tag, tagVal))
				}
			}
		}

		return nil
	})
}
