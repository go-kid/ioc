package meta

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"reflect"
)

type Node struct {
	Source  *Source
	Field   reflect.StructField
	Tag     string
	TagVal  string
	Type    reflect.Type
	Value   reflect.Value
	Injects []*Meta
}

func (n *Node) Id() string {
	return fmt.Sprintf("%s.%s[offset:%d]", n.Source.Type, n.Field.Name, n.Field.Offset)
}

func (n *Node) Name() string {
	if n.TagVal != "" {
		return n.TagVal
	}
	return GetComponentName(reflectx.New(n.Type))
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
	var c any
	switch t.(type) {
	case reflect.Value:
		c = t.(reflect.Value).Interface()
	default:
		c = t
	}
	if n, ok := c.(defination.NamingComponent); ok {
		return n.Naming()
	}
	return reflectx.Id(c)
}
