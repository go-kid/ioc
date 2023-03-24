package list

import (
	"sync"
)

type ConcurrentSets struct {
	cm sync.Map
}

func NewConcurrentSets(arr ...string) (s *ConcurrentSets) {
	s = &ConcurrentSets{}
	if arr != nil && len(arr) != 0 {
		s.PutAll(arr...)
	}
	return
}

func (r *ConcurrentSets) Put(s string) {
	r.cm.Store(s, struct{}{})
}

func (r *ConcurrentSets) PutAll(arr ...string) {
	for _, s := range arr {
		r.Put(s)
	}
}

func (r *ConcurrentSets) ToArray() (arr []string) {
	r.ForEach(func(k string) {
		arr = append(arr, k)
	})
	return
}

func (r *ConcurrentSets) Exists(t string) (ok bool) {
	_, ok = r.cm.Load(t)
	return ok
}

func (r *ConcurrentSets) ExistsAny(t ...string) (ok bool) {
	for _, s := range t {
		if r.Exists(s) {
			return true
		}
	}
	return false
}

func (r *ConcurrentSets) ExistsAll(t ...string) (ok bool) {
	for _, s := range t {
		if !r.Exists(s) {
			return false
		}
	}
	return true
}

func (r *ConcurrentSets) Remove(s string) {
	r.cm.Delete(s)
}

func (r *ConcurrentSets) RemoveAll(arr ...string) {
	for _, s := range arr {
		r.Remove(s)
	}
}

func (r *ConcurrentSets) Length() int {
	return len(r.ToArray())
}

func (r *ConcurrentSets) ForEach(accept func(key string)) {
	r.cm.Range(func(k interface{}, v interface{}) bool {
		accept(k.(string))
		return true
	})
}
