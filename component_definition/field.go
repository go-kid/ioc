package component_definition

import (
	"fmt"
	"reflect"
)

type Field struct {
	*Base
	Holder      *Holder
	StructField reflect.StructField
}

func (f *Field) ID() string {
	return fmt.Sprintf("%s.Field(%s)", f.Holder.ID(), f.StructField.Name)
}
