package meta

import (
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"reflect"
)

type Node struct {
	Field   reflect.StructField
	Tag     string
	TagVal  string
	Type    reflect.Type
	Value   reflect.Value
	Injects []*Meta
}

func (n *Node) Id() string {
	if n.TagVal != "" {
		return n.TagVal
	}
	return GetComponentName(n.Value)
}

func (n *Node) Inject(m ...*Meta) {
	if len(m) < 1 {
		return
	}
	switch n.Type.Kind() {
	case reflect.Slice:
		values := lo.Map(m, func(item *Meta, _ int) reflect.Value {
			return item.Value
		})
		n.Value.Set(reflect.Append(n.Value, values...))
	default:
		n.Value.Set(m[0].Value)
	}
	n.Injects = m
}

func GetComponentName(t any) string {
	switch t.(type) {
	case reflect.Value:
		c := t.(reflect.Value).Interface()
		if n, ok := c.(defination.NamingComponent); ok {
			return n.Naming()
		}
		return reflectx.Id(c)
	default:
		return reflectx.Id(t)
	}
}
