package scanner

import (
	"github.com/go-kid/ioc/scanner/meta"
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
		new(scanConfigurationPolicy),
		new(scanComponentPolicy),
	)
}

func (s *scanner) AddScanPolicies(policies ...ScanPolicy) {
	s.policies = append(s.policies, policies...)
}

func (s *scanner) ScanComponent(c any) *meta.Meta {
	m := meta.NewMeta(c)
	s.scanNodes(meta.NewHolder(m))
	return m
}

func (s *scanner) scanNodes(holder *meta.Holder) {
	_ = reflectx.ForEachFieldV2(holder.Type, holder.Value, false, func(field reflect.StructField, value reflect.Value) error {
		var base = &meta.Base{
			Type:  field.Type,
			Value: value,
		}
		//if is embed struct, find inside
		if field.Anonymous && field.Tag == "" && field.Type.Kind() == reflect.Struct {
			s.scanNodes(meta.NewEmbedHolder(base, holder))
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
					holder.Meta.SetNodes(nt, meta.NewNode(base, holder, field, tag, tagVal))
					continue
				}
			}

			//if not find in tag, use extract tag handler
			if handler := sp.ExtHandler(); handler != nil {
				if tag, tagVal, ok := handler(field, value); ok {
					holder.Meta.SetNodes(nt, meta.NewNode(base, holder, field, tag, tagVal))
				}
			}
		}

		return nil
	})
}
