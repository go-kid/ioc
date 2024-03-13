package meta

import (
	"fmt"
	"github.com/go-kid/ioc/syslog"
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
}

func NewNode(base *Base, holder *Holder, field reflect.StructField, tag, tagVal string) *Node {
	return &Node{
		Base:   base,
		Holder: holder,
		Field:  field,
		Tag:    tag,
		TagVal: tagVal,
	}
}

func (n *Node) ID() string {
	return fmt.Sprintf("%s.Field(%s)", n.Holder.ID(), n.Field.Name)
}

func (n *Node) Inject(metas ...*Meta) error {
	if len(metas) < 1 {
		return nil
	} else {
		var filteredMetas = make([]*Meta, 0, len(metas))
		for _, m := range metas {
			if m.ID() != n.Holder.Meta.ID() {
				filteredMetas = append(filteredMetas, m)
			}
		}
		if len(filteredMetas) == 0 {
			var embedSb = strings.Builder{}
			_ = n.Holder.Walk(func(source *Holder) error {
				embedSb.WriteString("\n depended on " + source.ID())
				return nil
			})
			return fmt.Errorf("field %s %s: self inject not allowed", n.ID(), embedSb.String())
		}
		metas = filteredMetas
	}

	switch n.Type.Kind() {
	case reflect.Slice:
		n.Value.Set(reflect.MakeSlice(n.Type, len(metas), len(metas)))
		for i, m := range metas {
			n.Value.Index(i).Set(m.Value)
		}
	default:
		if len(metas) > 1 {
			syslog.Warnf("inject multiple instances to single receiver %s, randomly select %s",
				n.ID(), metas[0].ID())
		}
		n.Value.Set(metas[0].Value)
	}

	for _, inject := range metas {
		inject.dependOn(n.Holder.Meta)
	}
	n.Injects = metas
	return nil
}
