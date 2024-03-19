package sync2

import "sync"

type Map[K, V any] struct {
	m sync.Map
}

func New[K, V any]() *Map[K, V] {
	return &Map[K, V]{}
}

func (m *Map[K, V]) Load(key K) (v V, loaded bool) {
	value, ok := m.m.Load(key)
	if ok {
		v = value.(V)
	}
	loaded = ok
	return
}

func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (V, bool) {
	actual, loaded := m.m.LoadOrStore(key, value)
	return actual.(V), loaded
}

func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *Map[K, V]) Range(f func(key K, value V) (shouldContinue bool)) {
	m.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}
