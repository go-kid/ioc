package list

import (
	"sync"
)

type gcset[T any] struct {
	cm sync.Map
}

func NewGenericConcurrentSets[T any](arr ...T) (s GenericSet[T]) {
	s = &gcset[T]{}
	if arr != nil && len(arr) != 0 {
		s.PutAll(arr...)
	}
	return
}

func (r *gcset[T]) Put(s T) {
	r.cm.Store(s, struct{}{})
}

func (r *gcset[T]) PutAll(arr ...T) {
	for _, s := range arr {
		r.Put(s)
	}
}

func (r *gcset[T]) ToArray() (arr []T) {
	r.ForEach(func(k T) {
		arr = append(arr, k)
	})
	return
}

func (r *gcset[T]) Exists(t T) (ok bool) {
	_, ok = r.cm.Load(t)
	return ok
}

func (r *gcset[T]) ExistsAny(t ...T) (ok bool) {
	for _, s := range t {
		if r.Exists(s) {
			return true
		}
	}
	return false
}

func (r *gcset[T]) ExistsAll(t ...T) (ok bool) {
	for _, s := range t {
		if !r.Exists(s) {
			return false
		}
	}
	return true
}

func (r *gcset[T]) Remove(s T) {
	r.cm.Delete(s)
}

func (r *gcset[T]) RemoveAll(arr ...T) {
	for _, s := range arr {
		r.Remove(s)
	}
}

func (r *gcset[T]) Length() int {
	return len(r.ToArray())
}

func (r *gcset[T]) ForEach(accept func(key T)) {
	r.cm.Range(func(k any, v any) bool {
		accept(k.(T))
		return true
	})
}
