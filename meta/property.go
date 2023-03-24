package meta

import "reflect"

type property struct {
	Prefix string
	Type   reflect.Type
	Value  reflect.Value
}
