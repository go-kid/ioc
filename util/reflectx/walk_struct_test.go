package reflectx

import (
	"fmt"
	"github.com/samber/lo"
	"reflect"
	"testing"
	"time"
)

type T struct {
	TEmbed
	Element    TElement
	TF1        string
	TComponent *TComponent
}

type TEmbed struct {
	TEbF1 int
}

type TElement struct {
	TElF1 float64
}

type TComponent struct {
	TCF1 time.Time
	Sub  *TComponent
}

func TestWalkField(t *testing.T) {
	WalkField(&T{
		TComponent: &TComponent{},
	}, func(parent *Node, field reflect.StructField, value reflect.Value) error {
		if parent == nil {
			fmt.Println(field.Name)
		} else {
			fmt.Println(lo.Map(parent.Path(), func(item *Node, _ int) string {
				return item.Field.Name
			}), field.Name)
		}
		if field.Name == "Sub" {
			println(value.IsValid())
		}
		return nil
	})
}
