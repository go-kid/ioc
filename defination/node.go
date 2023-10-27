package defination

import (
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
	//return GetComponentName(n.Type)
	if v := n.Value.Interface(); v != nil {
		return GetComponentName(v)
	}
	return GetComponentName(reflect.New(n.Type).Interface())
}

func GetComponentName(c interface{}) string {
	//fmt.Println(reflectx.IsTypeImplement(reflect.TypeOf(c), new(NamingComponent)))
	if n, ok := c.(NamingComponent); ok {
		return n.Naming()
	}
	t := reflect.TypeOf(c)
	//if t.Kind() == reflect.Ptr {
	//	t = t.Elem()
	//}
	return reflectx.TypeId(t)
}

func GetComponentNameByType(c reflect.Type) string {
	if reflectx.IsTypeImplement(c, new(NamingComponent)) {
		return reflectx.New(c).Interface().(NamingComponent).Naming()
	}
	return reflectx.TypeId(c)
}
