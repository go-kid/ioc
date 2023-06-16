package list

type gset[T comparable] struct {
	sets map[T]struct{}
}

func NewGenericSets[T comparable](arr ...T) GenericSet[T] {
	var s = &gset[T]{
		sets: make(map[T]struct{}),
	}
	s.PutAll(arr...)
	return s
}

func (r *gset[T]) Put(s T) {
	r.sets[s] = struct{}{}
}

func (r *gset[T]) PutAll(arr ...T) {
	for _, s := range arr {
		r.Put(s)
	}
}

func (r *gset[T]) ToArray() (arr []T) {
	r.ForEach(func(key T) {
		arr = append(arr, key)
	})
	return
}

func (r *gset[T]) Exists(t T) (ok bool) {
	_, ok = r.sets[t]
	return
}

func (r *gset[T]) ExistsAny(t ...T) (ok bool) {
	for _, s := range t {
		if r.Exists(s) {
			return true
		}
	}
	return false
}

func (r *gset[T]) ExistsAll(t ...T) (ok bool) {
	for _, s := range t {
		if !r.Exists(s) {
			return false
		}
	}
	return true
}

func (r *gset[T]) Remove(s T) {
	delete(r.sets, s)
}

func (r *gset[T]) RemoveAll(arr ...T) {
	for _, s := range arr {
		r.Remove(s)
	}
}

func (r *gset[T]) Length() int {
	return len(r.sets)
}

func (r *gset[T]) ForEach(accept func(key T)) {
	for k, _ := range r.sets {
		accept(k)
	}
}
