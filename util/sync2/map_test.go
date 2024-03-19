package sync2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap_Store(t *testing.T) {

	t.Run("TestLoad", func(t *testing.T) {
		m := New[string, data]()
		m.Store("1", data{"a"})
		m.Store("2", data{"b"})
		m.Store("3", data{"c"})
		load, ok := m.Load("1")
		assert.True(t, ok)
		assert.Equal(t, "a", load.v)

		load, ok = m.Load("2")
		assert.True(t, ok)
		assert.Equal(t, "b", load.v)

		load, ok = m.Load("3")
		assert.True(t, ok)
		assert.Equal(t, "c", load.v)

		load, ok = m.Load("4")
		assert.False(t, ok)
	})
	t.Run("LoadOrStore", func(t *testing.T) {
		m := New[string, data]()
		m.Store("1", data{"a"})
		m.Store("2", data{"b"})
		m.Store("3", data{"c"})
		load, ok := m.LoadOrStore("1", data{"a1"})
		assert.True(t, ok)
		assert.Equal(t, "a", load.v)

		load, ok = m.LoadOrStore("2", data{"b1"})
		assert.True(t, ok)
		assert.Equal(t, "b", load.v)

		load, ok = m.LoadOrStore("3", data{"c1"})
		assert.True(t, ok)
		assert.Equal(t, "c", load.v)

		load, ok = m.LoadOrStore("4", data{"d1"})
		assert.False(t, ok)
		assert.Equal(t, "d1", load.v)
	})
	t.Run("Delete", func(t *testing.T) {
		m := New[string, data]()
		m.Store("1", data{"a"})
		m.Store("2", data{"b"})
		m.Store("3", data{"c"})
		m.Delete("1")
		load, ok := m.Load("1")
		assert.False(t, ok)
		assert.Equal(t, "", load.v)

		m.Delete("10")
	})
	t.Run("Range", func(t *testing.T) {
		m := New[string, data]()
		m.Store("1", data{"a"})
		m.Store("2", data{"b"})
		m.Store("3", data{"c"})
		m.Range(func(key string, value data) (shouldContinue bool) {
			assert.True(t, key == "1" || key == "2" || key == "3")
			assert.True(t, value.v == "a" || value.v == "b" || value.v == "c")
			return true
		})
	})
}

type data struct {
	v string
}
