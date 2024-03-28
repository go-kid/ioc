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

func (b *Base) Update(c any) {
	b.Type = reflect.TypeOf(c)
	b.Value = reflect.ValueOf(c)
}
