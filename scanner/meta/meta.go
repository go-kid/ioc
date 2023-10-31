package meta

import (
	"fmt"
	"github.com/samber/lo"
	"reflect"
)

const (
	InjectTag  = "wire"
	ProduceTag = "produce"
	PropTag    = "prop"
)

type Meta struct {
	Name      string
	Address   string
	Raw       interface{}
	Type      reflect.Type
	Value     reflect.Value
	Produce   []*Meta
	DependsBy []*Meta

	Dependencies    []*Node
	Properties      []*Node
	CustomizedField []*Node
}

func (m *Meta) ID() string {
	return fmt.Sprintf("%s(%s#%s)", m.Name, m.Type, m.Address)
}

func (m *Meta) DependBy(parent *Meta) {
	if !lo.ContainsBy(m.DependsBy, func(item *Meta) bool {
		return item.ID() == parent.ID()
	}) {
		m.DependsBy = append(m.DependsBy, parent)
	}
}

func (m *Meta) AllDependencies() []*Node {
	return append(m.Dependencies, m.CustomizedField...)
}

func StringEscape(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}
