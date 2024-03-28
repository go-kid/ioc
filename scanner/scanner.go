package scanner

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
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
	m.SetNodes(s.scanNodes(m)...)
	return m
}

func (s *scanner) scanNodes(m *component_definition.Meta) []*component_definition.Node {
	var nodes []*component_definition.Node
	for _, field := range m.Fields {
		for _, sp := range s.policies {
			nodeType := sp.Group()
			if tag := sp.Tag(); tag != "" {
				if tagVal, ok := field.StructField.Tag.Lookup(tag); ok {
					node := component_definition.NewNode(field, nodeType, tag, tagVal)
					nodes = append(nodes, node)
					continue
				}
			}
			if handler := sp.ExtHandler(); handler != nil {
				if tag, tagVal, ok := handler(field.StructField, field.Value); ok {
					node := component_definition.NewNode(field, nodeType, tag, tagVal)
					nodes = append(nodes, node)
				}
			}
		}
	}
	return nodes
}
