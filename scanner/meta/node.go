package meta

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/util/reflectx"
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

func (n *Node) ID() string {
	return fmt.Sprintf("%s.Field(%s).Tag(%s)", n.Holder.ID(), n.Field.Name, n.Tag)
}

var (
	primaryInterface = new(defination.WirePrimary)
)

func (n *Node) Inject(metas []*Meta) error {
	var (
		isRequired = n.args.Has(ArgRequired, "true")
	)
	filtered, err := n.injectFilter(metas)
	if err != nil {
		if len(filtered) == 0 {
			if isRequired {
				return err
			}
			return nil
		}
		return err
	}
	metas = filtered

	switch n.Type.Kind() {
	case reflect.Slice:
		n.Value.Set(reflect.MakeSlice(n.Type, len(metas), len(metas)))
		for i, m := range metas {
			n.Value.Index(i).Set(m.Value)
		}
	default:
		var candidate = metas[0]
		if len(metas) > 1 {
			for _, m := range metas {
				//Primary interface first
				if reflectx.IsTypeImplement(m.Type, primaryInterface) {
					candidate = m
					break
				}
				//non naming component is preferred in multiple candidates
				if !m.IsAlias {
					candidate = m
				}
			}
		}
		n.Value.Set(candidate.Value)
		metas = []*Meta{candidate}
	}

	for _, inject := range metas {
		inject.dependOn(n.Holder.Meta)
	}
	n.Injects = metas
	return nil
}

func (n *Node) injectFilter(metas []*Meta) ([]*Meta, error) {
	//remove nil meta
	result := filter(metas, func(m *Meta) bool {
		return m != nil
	})
	if len(result) == 0 {
		return nil, fmt.Errorf("%s not found available components", n.ID())
	}
	//remove self-inject
	result = filter(result, func(m *Meta) bool {
		return m.ID() != n.Holder.Meta.ID()
	})
	if len(result) == 0 {
		var embedSb = strings.Builder{}
		_ = n.Holder.Walk(func(source *Holder) error {
			embedSb.WriteString("\n depended on " + source.ID())
			return nil
		})
		return nil, fmt.Errorf("field %s %s: self inject not allowed", n.ID(), embedSb.String())
	}
	//filter qualifier
	qualifierName, isQualifier := n.args.Find(ArgQualifier)
	if isQualifier {
		result = filter(result, func(m *Meta) bool {
			qualifier, ok := m.Raw.(defination.WireQualifier)
			return ok && n.args.Has(ArgQualifier, qualifier.Qualifier())
		})
		if len(result) == 0 {
			return nil, fmt.Errorf("field %s: no component found for qualifier %s", n.ID(), qualifierName)
		}
	}
	return result, nil
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
