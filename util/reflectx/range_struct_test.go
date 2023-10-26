package reflectx

import (
	"reflect"
	"testing"
)

func TestForEachField(t *testing.T) {
	var s string
	ForEachField(s, true, func(field reflect.StructField, value reflect.Value) error {
		return nil
	})
}
