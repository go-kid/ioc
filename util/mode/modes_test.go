package mode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModeEq(t *testing.T) {
	combined := M1 | M3 | M20
	assert.True(t, combined.Eq(M1))
	assert.True(t, combined.Eq(M3|M4))
	assert.True(t, combined.Eq(M20))
	assert.False(t, combined.Eq(M2))
	assert.False(t, combined.Eq(0))
}
