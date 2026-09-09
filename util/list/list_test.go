package list

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetImplementations(t *testing.T) {
	factories := map[string]func(...string) Set{
		"map":        NewSets,
		"concurrent": NewConcurrentSets,
	}

	for name, factory := range factories {
		t.Run(name, func(t *testing.T) {
			set := factory("a", "b")
			set.Put("c")
			set.PutAll("c", "d")

			assert.Equal(t, 4, set.Length())
			assert.True(t, set.Exists("a"))
			assert.True(t, set.ExistsAny("missing", "b"))
			assert.True(t, set.ExistsAll("a", "b", "c"))
			assert.False(t, set.ExistsAll("a", "missing"))
			assert.ElementsMatch(t, []string{"a", "b", "c", "d"}, set.ToArray())

			visited := make([]string, 0, set.Length())
			set.ForEach(func(key string) { visited = append(visited, key) })
			assert.ElementsMatch(t, set.ToArray(), visited)

			set.Remove("a")
			set.RemoveAll("b", "missing")
			assert.ElementsMatch(t, []string{"c", "d"}, set.ToArray())
		})
	}
}

func TestGenericSetImplementations(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		assertGenericSetBehavior(t, NewGenericSets(1, 2))
	})
	t.Run("concurrent", func(t *testing.T) {
		assertGenericSetBehavior(t, NewGenericConcurrentSets(1, 2))
	})
}

func assertGenericSetBehavior(t *testing.T, set GenericSet[int]) {
	t.Helper()
	set.Put(3)
	set.PutAll(3, 4)
	assert.Equal(t, 4, set.Length())
	assert.True(t, set.Exists(1))
	assert.True(t, set.ExistsAny(0, 2))
	assert.True(t, set.ExistsAll(1, 2, 3))
	assert.False(t, set.ExistsAll(1, 9))
	assert.ElementsMatch(t, []int{1, 2, 3, 4}, set.ToArray())

	visited := make([]int, 0, set.Length())
	set.ForEach(func(key int) { visited = append(visited, key) })
	assert.ElementsMatch(t, set.ToArray(), visited)

	set.Remove(1)
	set.RemoveAll(2, 9)
	assert.ElementsMatch(t, []int{3, 4}, set.ToArray())
}

func TestConcurrentSetConcurrentAccess(t *testing.T) {
	set := NewGenericConcurrentSets[int]()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(value int) {
			defer wg.Done()
			set.Put(value)
			assert.True(t, set.Exists(value))
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 100, set.Length())
}

func TestListOperations(t *testing.T) {
	values := []string{"a", "b", "c"}
	list := NewList(values)
	assert.Equal(t, len(values), list.Len())
	assert.True(t, list.Contains("b"))
	assert.False(t, list.Contains("missing"))
	index, ok := list.Find("b")
	require.True(t, ok)
	assert.Equal(t, 1, index)

	visited := make([]int, 0, len(values))
	list.ForEach(func(i int) { visited = append(visited, i) })
	assert.Equal(t, []int{0, 1, 2}, visited)

	visited = nil
	list.ForEachWithStop(func(i int) bool {
		visited = append(visited, i)
		return i == 1
	})
	assert.Equal(t, []int{0, 1}, visited)

	index, ok = list.FindBy(func(i int) bool { return values[i] == "missing" })
	assert.False(t, ok)
	assert.Equal(t, -1, index)
}
