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
	id      string
	Name    string
	IsAlias bool

	Raw    interface{}
	Fields []*Field

	dependedOnSet *sync2.Map[string, struct{}]
	DependedOn    []*Meta

	nodeGroup map[NodeType][]*Node
}

func NewMeta(c any) *Meta {
	base := NewBase(c)
	name, alias := GetComponentName(c)
	m := &Meta{
		Base:          base,
		id:            ComponentId(c),
		Name:          name,
		IsAlias:       alias,
		Raw:           c,
		dependedOnSet: sync2.New[string, struct{}](),
		nodeGroup:     make(map[NodeType][]*Node),
	}
	m.scanFields(NewHolder(m))
	return m
}

func (m *Meta) ID() string {
	return fmt.Sprintf("%s(0x%x)", m.id, m.Value.Pointer())
}

func (m *Meta) IsSelf(o *Meta) bool {
	return m.Value.Pointer() == o.originAddress
}

func (m *Meta) dependOn(parent *Meta) {
	m.dependedOnSet.LoadOrStore(parent.ID(), struct{}{})
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
