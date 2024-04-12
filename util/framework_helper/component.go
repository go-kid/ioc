package framework_helper

import (
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

func GetComponentNameWithAlias(t any) (name, alias string) {
	var c any
	switch t.(type) {
	case reflect.Value:
		c = t.(reflect.Value).Interface()
	case reflect.Type:
		c = reflect.New(t.(reflect.Type)).Interface()
	default:
		c = t
	}
	name = reflectx.Id(c)
	if n, ok := c.(definition.NamingComponent); ok {
		alias = n.Naming()
	}
	return
}

func GetComponentName(t any) string {
	name, alias := GetComponentNameWithAlias(t)
	if alias != "" {
		name = alias
	}
	return name
}
