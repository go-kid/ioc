package meta

import "reflect"

type Property struct {
	Prefix string
	Type   reflect.Type
	Value  reflect.Value
}
