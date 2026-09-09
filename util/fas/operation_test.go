package fas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTernaryOperations(t *testing.T) {
	assert.Equal(t, "yes", TernaryOp(true, "yes", "no"))
	assert.Equal(t, "no", TernaryOp(false, "yes", "no"))

	called := ""
	got := TernaryOpNil(false,
		func() int { called = "true"; return 1 },
		func() int { called = "false"; return 2 },
	)
	assert.Equal(t, 2, got)
	assert.Equal(t, "false", called)
}

func TestMinMaxAndFilter(t *testing.T) {
	assert.Equal(t, 5, Max(3, 5))
	assert.Equal(t, 3, Min(3, 5))
	assert.Equal(t, "b", Max("a", "b"))
	assert.Equal(t, []int{2, 4}, Filter([]int{1, 2, 3, 4}, func(v int) bool {
		return v%2 == 0
	}))
}

func TestIsNil(t *testing.T) {
	var (
		ptr   *int
		slice []string
		m     map[string]int
		ch    chan int
		fn    func()
	)

	for _, value := range []any{nil, ptr, slice, m, ch, fn} {
		assert.True(t, IsNil(value), "%T should be nil", value)
	}

	value := 1
	for _, nonNil := range []any{0, false, "", struct{}{}, &value, []int{}, map[string]int{}} {
		assert.False(t, IsNil(nonNil), "%T should not be nil", nonNil)
	}
}
