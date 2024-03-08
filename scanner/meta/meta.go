package meta

import (
	"fmt"
	"reflect"
)

const (
	InjectTag = "wire"
	PropTag   = "prop"
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
	Name    string
	Address string
	Raw     interface{}

	dependBySet map[string]struct{}
	DependsBy   []*Meta

	Dependencies    []*Node
	Properties      []*Node
	CustomizedField []*Node
}

func NewMeta(c any) *Meta {
	return &Meta{
		Base:        NewBase(c),
		Name:        GetComponentName(c),
		Address:     fmt.Sprintf("%p", c),
		Raw:         c,
		dependBySet: make(map[string]struct{}),
	}
}

func (m *Meta) ID() string {
	return fmt.Sprintf("%s.(%s)@%s", m.Name, m.Type, m.Address)
}

func (m *Meta) dependBy(parent *Meta) {
	if _, ok := m.dependBySet[parent.ID()]; !ok {
		m.DependsBy = append(m.DependsBy, parent)
		m.dependBySet[parent.ID()] = struct{}{}
	}
}

func (m *Meta) AllDependencies() []*Node {
	return append(m.Dependencies, m.CustomizedField...)
}
