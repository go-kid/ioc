package reflectx

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTypeId(t *testing.T) {
	WalkField(T{TComponent: &TComponent{}}, func(parent *Node, field reflect.StructField, value reflect.Value) error {
		fmt.Println(TypeId(field.Type))
		return nil
	})
}
