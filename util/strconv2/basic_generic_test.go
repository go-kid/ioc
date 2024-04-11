package strconv2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenericParse(t *testing.T) {
	t.Run("ParseString", func(t *testing.T) {
		parse, err := Parse[string]("abc")
		assert.NoError(t, err)
		assert.Equal(t, "abc", parse)
	})
	t.Run("ParseInt", func(t *testing.T) {
		i, err := Parse[int]("123")
		assert.NoError(t, err)
		assert.Equal(t, 123, i)
		i8, err := Parse[int8]("2")
		assert.NoError(t, err)
		assert.Equal(t, int8(2), i8)
		ui, err := Parse[uint]("123")
		assert.NoError(t, err)
		assert.Equal(t, uint(123), ui)
		ui8, err := Parse[uint8]("123")
		assert.NoError(t, err)
		assert.Equal(t, uint8(123), ui8)
	})
	t.Run("ParseFloat", func(t *testing.T) {
		f64, err := Parse[float64]("123.456")
		assert.NoError(t, err)
		assert.Equal(t, 123.456, f64)
		f32, err := Parse[float32]("123.456")
		assert.NoError(t, err)
		assert.Equal(t, float32(123.456), f32)
	})
	t.Run("ParseBool", func(t *testing.T) {
		parse, err := Parse[bool]("true")
		assert.NoError(t, err)
		assert.Equal(t, true, parse)
	})
	t.Run("ParseBool", func(t *testing.T) {
		parse, err := Parse[bool]("false")
		assert.NoError(t, err)
		assert.Equal(t, false, parse)
	})
}
