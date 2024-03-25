package component_definition

import (
	"fmt"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

func getComponentNameWithAlias(t any) (name, alias string) {
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

func GetComponentName(t any) (name string, isAlias bool) {
	var alias string
	name, alias = getComponentNameWithAlias(t)
	isAlias = alias != ""
	return
}

func ComponentId(t any) string {
	name, alias := getComponentNameWithAlias(t)
	if alias != "" {
		return fmt.Sprintf("%s(alias='%s')", name, alias)
	} else {
		return name
	}
}
