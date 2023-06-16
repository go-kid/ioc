package list

type set map[string]struct{}

func NewSets(arr ...string) (s Set) {
	s = make(set)
	s.PutAll(arr...)
	return
}

func (r set) Put(s string) {
	r[s] = struct{}{}
}

func (r set) PutAll(arr ...string) {
	for _, s := range arr {
		r.Put(s)
	}
}

func (r set) ToArray() (arr []string) {
	r.ForEach(func(key string) {
		arr = append(arr, key)
	})
	return
}

func (r set) Exists(t string) (ok bool) {
	_, ok = r[t]
	return
}

func (r set) ExistsAny(t ...string) (ok bool) {
	for _, s := range t {
		if r.Exists(s) {
			return true
		}
	}
	return false
}

func (r set) ExistsAll(t ...string) (ok bool) {
	for _, s := range t {
		if !r.Exists(s) {
			return false
		}
	}
	return true
}

func (r set) Remove(s string) {
	delete(r, s)
}

func (r set) RemoveAll(arr ...string) {
	for _, s := range arr {
		r.Remove(s)
	}
}

func (r set) Length() int {
	return len(r)
}

func (r set) ForEach(accept func(key string)) {
	for k, _ := range r {
		accept(k)
	}
}
