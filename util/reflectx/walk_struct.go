package reflectx

import (
	"github.com/samber/lo"
	"reflect"
)

type Node struct {
	Parent *Node
	Field  reflect.StructField
	Value  reflect.Value
}

func (n *Node) Path() []*Node {
	if n == nil {
		return nil
	}
	var parent = n
	var nodes []*Node
	for parent != nil {
		nodes = append(nodes, parent)
		parent = parent.Parent
	}
	return lo.Reverse(nodes)
}

type FieldWalkAcceptor func(parent *Node, field reflect.StructField, value reflect.Value) error

// WalkField walk the struct field with DFS, and break when meeting the nil pointer value, can be checked by value.Elem().InValid()
func WalkField(o interface{}, f FieldWalkAcceptor) error {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	return WalkFieldV2(t, v, f)
}

func WalkFieldV2(t reflect.Type, v reflect.Value, f FieldWalkAcceptor) error {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	return ForEachFieldV2(t, v, false, func(field reflect.StructField, value reflect.Value) error {
		err := walkField(nil, field, value, f)
		if err != nil {
			return err
		}
		return nil
	})
}

func walkField(node *Node, field reflect.StructField, value reflect.Value, f FieldWalkAcceptor) error {
	err := f(node, field, value)
	if err != nil {
		return err
	}
	t := field.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	if t.NumField() < 1 {
		return nil
	}
	for i := 0; i < t.NumField(); i++ {
		err := walkField(&Node{
			Parent: node,
			Field:  field,
			Value:  value,
		}, t.Field(i), value.Field(i), f)
		if err != nil {
			return err
		}
	}
	return nil
}
