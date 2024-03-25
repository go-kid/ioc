package scanner

import (
	"github.com/go-kid/ioc/definition"
	"reflect"
)

func propHandler(field reflect.StructField, value reflect.Value) (tag string, tagVal string, ok bool) {
	if configuration, infer := value.Interface().(definition.Configuration); infer {
		tag = definition.PropTag
		tagVal = configuration.Prefix()
		ok = true
	}
	return
}
