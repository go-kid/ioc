package component_definition

import (
	"fmt"
	"reflect"
)

type Property struct {
	*Field
	PropertyType PropertyType
	Tag          string
	TagVal       string
	Injects      []*Meta
	args         TagArg
}

func NewProperty(field *Field, propType PropertyType, tag, tagVal string) *Property {
	parsedTagVal, arg := defaultPropertyArgs().Parse(tagVal)
	return &Property{
		Field:        field,
		PropertyType: propType,
		Tag:          tag,
		TagVal:       parsedTagVal,
		args:         arg,
	}
}

func defaultPropertyArgs() TagArg {
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

func (n *Property) ID() string {
	return fmt.Sprintf("%s.Tag(%s:'%s').Type(%s)", n.Field.ID(), n.Tag, n.TagVal, n.PropertyType)
}

func (n *Property) String() string {
	return n.ID()
}

func (n *Property) Inject(metas []*Meta) error {
	isRequired := n.args.Has(ArgRequired, "true")
	if len(metas) == 0 {
		if isRequired {
			return fmt.Errorf("%s not found available components", n.ID())
		}
		return nil
	}

	//remove self-inject
	metas = filter(metas, func(m *Meta) bool {
		return !n.Holder.Meta.IsSelf(m)
	})
	if len(metas) == 0 {
		if isRequired {
			return fmt.Errorf("field %s %s: self inject not allowed", n.ID(), n.Holder.Stack())
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

func (n *Property) Args() TagArg {
	return n.args
}

func (n *Property) SetArg(t ArgType, val []string) {
	n.args[t] = val
}
func (n *Property) AppendArg(t ArgType, val []string) {
	n.args[t] = append(n.args[t], val...)
}

func (n *Property) SetArgs(a TagArg) {
	for argType, val := range a {
		n.SetArg(argType, val)
	}
}

func (n *Property) AppendArgs(a TagArg) {
	for argType, val := range a {
		n.AppendArg(argType, val)
	}
}
