package di

import (
	"github.com/go-kid/ioc/defination"
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
	if v := n.Value.Interface(); v != nil {
		return defination.GetComponentName(v)
	}
	return defination.GetComponentName(reflect.New(n.Type).Interface())
}
