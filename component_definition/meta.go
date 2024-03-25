package component_definition

import (
	"fmt"
	"github.com/go-kid/ioc/util/sync2"
	"reflect"
)

type NodeType string

const (
	NodeTypeConfiguration NodeType = "configuration"
	NodeTypeComponent     NodeType = "component"
)

type Base struct {
	Type          reflect.Type
	Value         reflect.Value
	OriginAddress uintptr
}

func NewBase(c any) *Base {
	v := reflect.ValueOf(c)
	return &Base{
		Type:          reflect.TypeOf(c),
		Value:         v,
		OriginAddress: v.Pointer(),
	}
}

func (b *Base) Update(c any) {
	b.Type = reflect.TypeOf(c)
	b.Value = reflect.ValueOf(c)
}

type Meta struct {
	*Base
	id      string
	Name    string
	IsAlias bool

	Raw interface{}

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
	return m
}

func (m *Meta) OriginID() string {
	return fmt.Sprintf("%s(0x%x)", m.id, m.OriginAddress)
}

func (m *Meta) ID() string {
	return fmt.Sprintf("%s(0x%x)", m.id, m.Value.Pointer())
}

func (m *Meta) dependOn(parent *Meta) {
	m.dependedOnSet.LoadOrStore(parent.ID(), struct{}{})
}

func (m *Meta) SetNodes(t NodeType, nodes ...*Node) {
	m.nodeGroup[t] = append(m.nodeGroup[t], nodes...)
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
