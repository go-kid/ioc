package component_definition

import (
	"fmt"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/ioc/util/sync2"
	"reflect"
)

type NodeType string

const (
	NodeTypeConfiguration NodeType = "configuration"
	NodeTypeComponent     NodeType = "component"
)

type Meta struct {
	*Base
	ProxyMeta *Meta
	name      string
	alias     string

	Raw    interface{}
	Fields []*Field

	dependentSet *sync2.Map[string, struct{}]
	Dependent    []*Meta

	nodeGroup map[NodeType][]*Node
}

func NewMeta(c any) *Meta {
	base := NewBase(c)
	name, alias := GetComponentNameWithAlias(c)
	m := &Meta{
		Base:         base,
		name:         name,
		alias:        alias,
		Raw:          c,
		dependentSet: sync2.New[string, struct{}](),
		nodeGroup:    make(map[NodeType][]*Node),
	}
	m.scanFields(NewHolder(m))
	return m
}

func (m *Meta) ID() string {
	if m.IsAlias() {
		return fmt.Sprintf("%s(alias=%s)(0x%x)", m.name, m.alias, m.Value.Pointer())
	}
	return fmt.Sprintf("%s(0x%x)", m.name, m.Value.Pointer())
}

func (m *Meta) Name() string {
	if m.IsAlias() {
		return m.alias
	}
	return m.name
}

func (m *Meta) SetName(name string) {
	if name != m.name {
		m.alias = name
	}
}

func (m *Meta) IsAlias() bool {
	return m.alias != ""
}

func (m *Meta) IsSelf(o *Meta) bool {
	return m.Value.Pointer() == o.originAddress
}

func (m *Meta) dependOn(dependent *Meta) {
	_, loaded := m.dependentSet.LoadOrStore(dependent.ID(), struct{}{})
	if !loaded {
		m.Dependent = append(m.Dependent, dependent)
	}
}

func (m *Meta) GetDependents() (names []string) {
	for _, meta := range m.Dependent {
		names = append(names, meta.Name())
	}
	return
}

func (m *Meta) SetNodes(nodes ...*Node) {
	for _, node := range nodes {
		m.nodeGroup[node.NodeType] = append(m.nodeGroup[node.NodeType], node)
	}
}

func (m *Meta) GetNodes(t NodeType) []*Node {
	return m.nodeGroup[t]
}

func (m *Meta) GetComponentNodes() []*Node {
	return m.GetNodes(NodeTypeComponent)
}

func (m *Meta) GetConfigurationNodes() []*Node {
	return m.GetNodes(NodeTypeConfiguration)
}

func (m *Meta) GetAllNodes() []*Node {
	var nodes []*Node
	for _, groupNodes := range m.nodeGroup {
		nodes = append(nodes, groupNodes...)
	}
	return nodes
}

func (m *Meta) UseProxy(origin any) {
	m.ProxyMeta = &Meta{
		Base:         m.Base,
		ProxyMeta:    nil,
		name:         m.name,
		alias:        m.alias,
		Raw:          m.Raw,
		Fields:       m.Fields,
		dependentSet: m.dependentSet,
		Dependent:    m.Dependent,
		nodeGroup:    m.nodeGroup,
	}
	m.Base = NewBase(origin)
	m.Raw = origin
}

func (m *Meta) scanFields(holder *Holder) {
	_ = reflectx.ForEachFieldV2(holder.Type, holder.Value, false, func(field reflect.StructField, value reflect.Value) error {
		var base = &Base{
			Type:  field.Type,
			Value: value,
		}
		//if is embed struct, find inside
		if field.Anonymous && field.Tag == "" && field.Type.Kind() == reflect.Struct {
			m.scanFields(NewEmbedHolder(base, holder))
			return nil
		}

		if !value.CanSet() {
			return nil
		}
		m.Fields = append(m.Fields, &Field{
			Base:        base,
			Holder:      holder,
			StructField: field,
		})
		return nil
	})
}

func (m *Meta) IsSingleton() bool {
	return true
}
