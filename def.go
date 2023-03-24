package ioc

import "reflect"

const (
	injectTag = "wire"
	configTag = "prop"
)

type InitializeComponent interface {
	Init() error
}

type NamingComponent interface {
	Naming() string
}

type ApplicationRunner interface {
	Order() int
	Run() error
}

type Configuration interface {
	Prefix() string
}

type ComponentPostProcessor interface {
	PostProcessBeforeInitialization(component interface{}) error
	PostProcessAfterInitialization(component interface{}) error
}

func isDependency(field reflect.StructField) (string, bool) {
	value, ok := field.Tag.Lookup(injectTag)
	return value, ok
}

func isConfigure(field reflect.StructField, value reflect.Value) (string, bool) {
	if key, ok := field.Tag.Lookup(configTag); ok {
		return key, true
	}
	if configuration, ok := value.Interface().(Configuration); ok {
		return configuration.Prefix(), true
	}
	return "", false
}

func getComponentName(c interface{}) string {
	if n, ok := c.(NamingComponent); ok {
		return n.Naming()
	}
	t := reflect.TypeOf(c)
	if t.Kind() == reflect.Ptr {
		return t.Elem().String()
	}
	return t.String()
}
