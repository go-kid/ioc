package list

type Sets map[string]struct{}

func NewSets(arr ...string) (s Sets) {
	s = make(Sets)
	s.PutAll(arr...)
	return
}

func (r Sets) Put(s string) {
	r[s] = struct{}{}
}

func (r Sets) PutAll(arr ...string) {
	for _, s := range arr {
		r.Put(s)
	}
}

func (r Sets) ToArray() (arr []string) {
	for k, _ := range r {
		arr = append(arr, k)
	}
	return
}

func (r Sets) Exists(t string) (ok bool) {
	_, ok = r[t]
	return
}

func (r Sets) ExistsAny(t ...string) (ok bool) {
	for _, s := range t {
		if r.Exists(s) {
			return true
		}
	}
	return false
}

func (r Sets) ExistsAll(t ...string) (ok bool) {
	for _, s := range t {
		if !r.Exists(s) {
			return false
		}
	}
	return true
}

func (r Sets) Remove(s string) {
	delete(r, s)
}

func (r Sets) RemoveAll(arr ...string) {
	for _, s := range arr {
		r.Remove(s)
	}
}

func (r Sets) Length() int {
	return len(r)
}

func (r Sets) ForEach(accept func(key string)) {
	for k, _ := range r {
		accept(k)
	}
}
