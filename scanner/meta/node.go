package meta

import (
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

type Node struct {
	Tag   string
	Type  reflect.Type
	Value reflect.Value
}

func (n *Node) Id() string {
	if n.Tag != "" {
		return n.Tag
	}
	return GetComponentName(n.Value)
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
