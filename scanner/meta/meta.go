package meta

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/util/fas"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/ioc/util/sync2"
	"reflect"
)

type NodeType string

const (
	NodeTypeConfiguration NodeType = "configuration"
	NodeTypeComponent     NodeType = "component"
)

type Base struct {
	Type  reflect.Type
	Value reflect.Value
}

func NewBase(c any) *Base {
	return &Base{
		Type:  reflect.TypeOf(c),
		Value: reflect.ValueOf(c),
	}
}

type Meta struct {
	*Base
	typeId  string
	fullId  string
	Name    string
	IsAlias bool
	Address uintptr
	Raw     interface{}

	dependedOnSet *sync2.Map[string, struct{}]
	DependedOn    []*Meta

	nodeGroup map[NodeType][]*Node
}

func NewMeta(c any) *Meta {
	base := NewBase(c)
	address := base.Value.Pointer()
	typeId, alias := GetComponentName(c)
	var (
		isAlias = alias != ""
		name    = fas.TernaryOp(isAlias, alias, typeId)
		fullId  string
	)
	if isAlias {
		fullId = fmt.Sprintf("%s(alias='%s')(0x%x)", typeId, name, address)
	} else {
		fullId = fmt.Sprintf("%s(0x%x)", name, address)
	}
	m := &Meta{
		Base:          base,
		typeId:        typeId,
		fullId:        fullId,
		Name:          name,
		IsAlias:       isAlias,
		Address:       address,
		Raw:           c,
		dependedOnSet: sync2.New[string, struct{}](),
		nodeGroup:     make(map[NodeType][]*Node),
	}
	return m
}

func (m *Meta) ID() string {
	return m.fullId
}

func (m *Meta) TypeID() string {
	return m.typeId
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

func GetComponentName(t any) (id, alias string) {
	var c any
	switch t.(type) {
	case reflect.Value:
		c = t.(reflect.Value).Interface()
	case reflect.Type:
		c = reflect.New(t.(reflect.Type)).Interface()
	default:
		c = t
	}
	id = reflectx.Id(c)
	if n, ok := c.(defination.NamingComponent); ok {
		alias = n.Naming()
	}
	return
}
