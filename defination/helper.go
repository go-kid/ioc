package defination

import "reflect"

const (
	injectTag  = "wire"
	produceTag = "produce"
	configTag  = "prop"
)

func IsDependency(field reflect.StructField) (string, bool) {
	value, ok := field.Tag.Lookup(injectTag)
	return value, ok
}

func IsProduce(field reflect.StructField) (string, bool) {
	value, ok := field.Tag.Lookup(produceTag)
	return value, ok
}

func IsConfigure(field reflect.StructField, value reflect.Value) (string, bool) {
	if key, ok := field.Tag.Lookup(configTag); ok {
		return key, true
	}
	if configuration, ok := value.Interface().(Configuration); ok {
		return configuration.Prefix(), true
	}
	return "", false
}

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
