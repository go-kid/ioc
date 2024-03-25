package component_definition

import (
	"fmt"
	"reflect"
	"strings"
)

type Node struct {
	*Base
	Holder  *Holder
	Field   reflect.StructField
	Tag     string
	TagVal  string
	Injects []*Meta
	args    TagArg
}

func NewNode(base *Base, holder *Holder, field reflect.StructField, tag, tagVal string) *Node {
	parsedTagVal, arg := defaultNodeArgs().Parse(tagVal)
	return &Node{
		Base:    base,
		Holder:  holder,
		Field:   field,
		Tag:     tag,
		TagVal:  parsedTagVal,
		Injects: nil,
		args:    arg,
	}
}

func defaultNodeArgs() TagArg {
	return TagArg{
		ArgRequired: {"true"},
	}
}

func filter(metas []*Meta, f func(m *Meta) bool) []*Meta {
	var result = make([]*Meta, 0, len(metas))
	for _, m := range metas {
		if f(m) {
			result = append(result, m)
		}
	}
	return result
}

func (n *Node) ID() string {
	return fmt.Sprintf("%s.Field(%s).Tag(%s)", n.Holder.ID(), n.Field.Name, n.Tag)
}

func (n *Node) Inject(metas []*Meta) error {
	isRequired := n.args.Has(ArgRequired, "true")
	if len(metas) == 0 {
		if isRequired {
			return fmt.Errorf("%s not found available components", n.ID())
		}
		return nil
	}

	//remove self-inject
	metas = filter(metas, func(m *Meta) bool {
		return m.ID() != n.Holder.Meta.OriginID()
	})
	if len(metas) == 0 {
		if isRequired {
			var embedSb = strings.Builder{}
			_ = n.Holder.Walk(func(source *Holder) error {
				embedSb.WriteString("\n depended on " + source.ID())
				return nil
			})
			return fmt.Errorf("field %s %s: self inject not allowed", n.ID(), embedSb.String())
		}
		return nil
	}

	switch n.Type.Kind() {
	case reflect.Slice, reflect.Array:
		n.Value.Set(reflect.MakeSlice(n.Type, len(metas), len(metas)))
		for i, m := range metas {
			n.Value.Index(i).Set(m.Value)
			m.dependOn(n.Holder.Meta)
		}
	default:
		m := metas[0]
		n.Value.Set(m.Value)
		m.dependOn(n.Holder.Meta)
	}

	n.Injects = metas
	return nil
}

func (n *Node) Args() TagArg {
	return n.args
}

func (n *Node) SetArg(t ArgType, val []string) {
	n.args[t] = val
}
func (n *Node) AppendArg(t ArgType, val []string) {
	n.args[t] = append(n.args[t], val...)
}

func (n *Node) SetArgs(a TagArg) {
	for argType, val := range a {
		n.SetArg(argType, val)
	}
}

func (n *Node) AppendArgs(a TagArg) {
	for argType, val := range a {
		n.AppendArg(argType, val)
	}
}
