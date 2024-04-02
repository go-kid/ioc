package component_definition

import "reflect"

type Base struct {
	Type          reflect.Type
	Value         reflect.Value
	originAddress uintptr
}

func NewBase(c any) *Base {
	v := reflect.ValueOf(c)
	return &Base{
		Type:          reflect.TypeOf(c),
		Value:         v,
		originAddress: v.Pointer(),
	}
}
