package list

import (
	"reflect"
)

type List struct {
	inner interface{}
	v     reflect.Value
}

func NewList(slice interface{}) *List {
	return &List{
		inner: slice,
		v:     reflect.ValueOf(slice),
	}
}

func (r *List) value() reflect.Value {
	return r.v
}

func (r *List) ForEach(accept func(i int)) {
	for i := 0; i < r.Len(); i++ {
		accept(i)
	}
}

func (r *List) ForEachWithStop(accept func(i int) bool) {
	for i := 0; i < r.Len(); i++ {
		if accept(i) {
			break
		}
	}
}

func (r *List) FindBy(accept func(i int) bool) (j int, ok bool) {
	r.ForEachWithStop(func(i int) bool {
		j = i
		ok = accept(i)
		return ok
	})
	if !ok {
		j = -1
	}
	return
}

func (r *List) Find(o interface{}) (int, bool) {
	return r.FindBy(func(i int) bool {
		return r.value().Index(i).Interface() == o
	})
}

func (r *List) Contains(o interface{}) (ok bool) {
	_, ok = r.Find(o)
	return
}

func (r *List) Len() int {
	return r.value().Len()
}
