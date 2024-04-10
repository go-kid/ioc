package reflectx

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestSetAnyValueFromString(t *testing.T) {
	type ST struct {
		A string
		B *ST
	}
	type T struct {
		S      string           `val:"hello"`
		SE     string           `val:"\"hello\""`
		B      bool             `val:"true"`
		I      int32            `val:"32"`
		F      float32          `val:"2.33"`
		Slice  []uint16         `val:"[1,2,3]"`
		Array  [3]uint32        `val:"[3,2,1]"`
		PS     *string          `val:"hello"`
		M      map[string]any   `val:"map[a:0 b:1 c:2]"`
		SliceM []map[string]any `val:"[map[a:1],map[b:2]]"`
	}
	var tt = &T{}
	var expected = map[string]any{
		"S":     "hello",
		"SE":    "hello",
		"B":     true,
		"I":     int32(32),
		"F":     float32(2.33),
		"Slice": []uint16{1, 2, 3},
		"Array": [3]uint32{3, 2, 1},
		"PS": func() *string {
			var s = "hello"
			return &s
		}(),
		"M": map[string]any{
			"a": int64(0),
			"b": int64(1),
			"c": int64(2),
		},
		"SliceM": []map[string]any{
			{
				"a": int64(1),
			},
			{
				"b": int64(2),
			},
		},
	}
	err := ForEachField(tt, true, func(field reflect.StructField, value reflect.Value) error {
		err := SetAnyValueFromString(field.Type, value, field.Tag.Get("val"))
		if err != nil {
			return err
		}
		assert.Equal(t, expected[field.Name], value.Interface())
		return nil
	})
	assert.NoError(t, err)
}
