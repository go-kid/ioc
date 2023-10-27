package injector

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

func GetComponentName(t reflect.Value) string {
	if n, ok := t.Interface().(defination.NamingComponent); ok {
		return n.Naming()
	}
	return reflectx.Id(t.Interface())
}
