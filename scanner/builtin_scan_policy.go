package scanner

import (
	"github.com/go-kid/ioc/defination"
	"reflect"
)

func propHandler(field reflect.StructField, value reflect.Value) (tag string, tagVal string, ok bool) {
	if configuration, infer := value.Interface().(defination.Configuration); infer {
		tag = defination.PropTag
		tagVal = configuration.Prefix()
		ok = true
	}
	return
}
