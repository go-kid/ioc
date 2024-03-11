package meta

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"reflect"
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
	return fmt.Sprintf("%s.%s", n.Holder.ID(), n.Field.Name)
}

func (n *Node) Name() string {
	if n.TagVal != "" {
		return n.TagVal
	}
	return GetComponentName(reflectx.New(n.Type))
}

func (n *Node) Inject(m ...*Meta) error {
	if len(m) < 1 {
		return nil
	}
	var value reflect.Value
	switch n.Type.Kind() {
	case reflect.Slice:
		values := lo.FilterMap(m, func(item *Meta, _ int) (reflect.Value, bool) {
			if m[0].ID() == n.Holder.Meta.ID() {
				return reflect.Value{}, false
			}
			return item.Value, true
		})
		value = reflect.Append(n.Value, values...)
	default:
		if m[0].ID() == n.Holder.Meta.ID() {
			return fmt.Errorf("self inject is not allowed: %s", m[0].ID())
		}
		value = m[0].Value
	}
	n.Value.Set(value)

	for _, inject := range m {
		inject.dependBy(n.Holder.Meta)
	}
	n.Injects = m
	return nil
}

func GetComponentName(t any) string {
	var c any
	switch t.(type) {
	case reflect.Value:
		c = t.(reflect.Value).Interface()
	default:
		c = t
	}
	if n, ok := c.(defination.NamingComponent); ok && n.Naming() != "" {
		return n.Naming()
	}
	return reflectx.Id(c)
}
