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
	return &Node{
		Base:    base,
		Holder:  holder,
		Field:   field,
		Tag:     tag,
		TagVal:  tagVal,
		Injects: nil,
		args:    defaultNodeArgs().Parse(tagVal),
	}
}

func defaultNodeArgs() TagArg {
	return TagArg{
		ArgRequired:  {"true"},
		ArgQualifier: nil,
	}
}
func (n *Node) ID() string {
	return fmt.Sprintf("%s.Field(%s)", n.Holder.ID(), n.Field.Name)
}

var (
	qualifierInterface = new(defination.WireQualifier)
	primaryInterface   = new(defination.WirePrimary)
)

func (n *Node) Inject(metas ...*Meta) error {
	required := n.args.Has(ArgRequired, "true")
	if len(metas) == 0 {
		if required {
			return fmt.Errorf("%s inject null components", n.ID())
		}
		return nil
	}
	//filter self-inject
	var filteredMetas = make([]*Meta, 0, len(metas))
	for _, m := range metas {
		if m.ID() != n.Holder.Meta.ID() {
			filteredMetas = append(filteredMetas, m)
		}
	}
	if len(filteredMetas) == 0 {
		if required {
			var embedSb = strings.Builder{}
			_ = n.Holder.Walk(func(source *Holder) error {
				embedSb.WriteString("\n depended on " + source.ID())
				return nil
			})
			return fmt.Errorf("field %s %s: self inject not allowed", n.ID(), embedSb.String())
		}
		return nil
	}

	metas = filteredMetas

	switch n.Type.Kind() {
	case reflect.Slice:
		n.Value.Set(reflect.MakeSlice(n.Type, len(metas), len(metas)))
		for i, m := range metas {
			n.Value.Index(i).Set(m.Value)
		}
	default:
		var candidate = metas[0]
		if len(metas) > 1 {
			_, isQualifier := n.args.Find(ArgQualifier)
			for _, m := range metas {
				if !m.IsAlias {
					candidate = m
				}
				if reflectx.IsTypeImplement(m.Type, primaryInterface) {
					candidate = m
				}
				if isQualifier && reflectx.IsTypeImplement(m.Type, qualifierInterface) {
					candidate = m
					break
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

func (n *Node) Args() TagArg {
	return n.args
}
