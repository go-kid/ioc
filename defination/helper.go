package defination

import "reflect"

func GetComponentName(c interface{}) string {
	if n, ok := c.(NamingComponent); ok {
		return n.Naming()
	}
	t := reflect.TypeOf(c)
	if t.Kind() == reflect.Ptr {
		return t.Elem().String()
	}
	return t.String()
}
