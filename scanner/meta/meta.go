package meta

import (
	"fmt"
	"github.com/samber/lo"
	"reflect"
	"strings"
)

//
//const (
//	InjectTag  = "wire"
//	ProduceTag = "produce"
//	PropTag    = "prop"
//)

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
	CustomizedField map[string][]*Node
}

func (m *Meta) ID() string {
	return fmt.Sprintf("%s(%s#%s)", m.Name, m.Type, m.Address)
}

func (m *Meta) DependBy(parent *Meta) {
	m.DependsBy = append(m.DependsBy, parent)
}

type kv struct {
	k string
	v string
}

func (m *Meta) DotNodeAttr() map[string]string {
	var label = []*kv{
		{k: "", v: m.Name},
		{k: "Type", v: m.Type.String()},
		{k: "Props", v: strings.Join(lo.Map[*Node, string](m.Properties, func(p *Node, _ int) string {
			return p.Tag
		}), ", ")},
	}

	labels := lo.Map[*kv, string](label, func(item *kv, _ int) string {
		if item.k == "" {
			return item.v
		}
		return fmt.Sprintf("%s: %s", item.k, item.v)
	})
	return map[string]string{
		"label": StringEscape("{" + strings.Join(labels, "|") + "}"),
		"shape": "record",
	}
}

func StringEscape(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}
