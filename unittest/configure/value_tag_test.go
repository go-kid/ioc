package configure

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValueTag(t *testing.T) {
	t.Run("TestSimpleType", func(t *testing.T) {
		t.Run("BaseType", func(t *testing.T) {
			type T struct {
				A string  `value:"foo"`
				B bool    `value:"true"`
				I int     `value:"123"`
				F float64 `value:"123.321"`
			}
			var tt = &T{}
			ioc.RunTest(t, app.SetComponents(tt))
			assert.Equal(t, "foo", tt.A)
			assert.True(t, tt.B)
			assert.Equal(t, 123, tt.I)
			assert.Equal(t, 123.321, tt.F)
		})
		t.Run("Slice", func(t *testing.T) {
			type T struct {
				S []string  `value:"[\"hello\",\"world\",\"foo\",\"bar\"]"`
				I []int     `value:"[1,2,3]"`
				B []bool    `value:"[true,false,false,true]"`
				F []float64 `value:"[1.1,2.2,3]"`
			}
			var tt = &T{}
			ioc.RunTest(t, app.SetComponents(tt))
			assert.Equal(t, []string{"hello", "world", "foo", "bar"}, tt.S)
			assert.Equal(t, []int{1, 2, 3}, tt.I)
			assert.Equal(t, []bool{true, false, false, true}, tt.B)
			assert.Equal(t, []float64{1.1, 2.2, 3}, tt.F)
		})
		t.Run("Map", func(t *testing.T) {
			type T struct {
				M  map[string]any `value:"{}"`
				ME map[string]any
				M2 map[string]any `value:"{\"foo\":\"bar\"}"`
			}
			var tt = &T{}
			ioc.RunTest(t, app.SetComponents(tt))
			assert.NotNil(t, tt.M)
			assert.Nil(t, tt.ME)
			assert.Equal(t, map[string]any{"foo": "bar"}, tt.M2)
		})
		t.Run("Struct", func(t *testing.T) {
			type S struct {
				Foo string `json:"foo"`
			}
			type T struct {
				M2 S `value:"{\"foo\":\"bar\"}"`
			}
			var tt = &T{}
			ioc.RunTest(t, app.SetComponents(tt))
			assert.Equal(t, S{Foo: "bar"}, tt.M2)
		})
	})
	t.Run("TestComplexType", func(t *testing.T) {
		t.Run("BasePointer", func(t *testing.T) {
			type T struct {
				Ap *string  `value:"foo"`
				B  *bool    `value:"true"`
				I  *int     `value:"123"`
				F  *float64 `value:"123.321"`
			}
			var tt = &T{}
			ioc.RunTest(t, app.SetComponents(tt))
			var foo = "foo"
			assert.Equal(t, &foo, tt.Ap)
			var b = true
			assert.Equal(t, &b, tt.B)
			var i = 123
			assert.Equal(t, &i, tt.I)
			var f = 123.321
			assert.Equal(t, &f, tt.F)
		})
		t.Run("SlicePointer", func(t *testing.T) {
			type T struct {
				S []*string  `value:"[\"hello\",\"world\",\"foo\",\"bar\"]"`
				I []*int     `value:"[1,2,3]"`
				B []*bool    `value:"[true,false,false,true]"`
				F []*float64 `value:"[1.1,2.2,3]"`
			}
			var tt = &T{}
			ioc.RunTest(t, app.SetComponents(tt))
			var ss = []string{"hello", "world", "foo", "bar"}
			for i, s := range ss {
				assert.Equal(t, &s, tt.S[i])
			}
			var is = []int{1, 2, 3}
			for i, s := range is {
				assert.Equal(t, &s, tt.I[i])
			}
			var bs = []bool{true, false, false, true}
			for i, s := range bs {
				assert.Equal(t, &s, tt.B[i])
			}
			var fs = []float64{1.1, 2.2, 3}
			for i, s := range fs {
				assert.Equal(t, &s, tt.F[i])
			}
		})
		t.Run("StructPointer", func(t *testing.T) {
			type S struct {
				Foo string `json:"foo"`
			}
			type T struct {
				M2 *S `value:"{\"foo\":\"bar\"}"`
			}
			var tt = &T{}
			ioc.RunTest(t, app.SetComponents(tt))
			assert.Equal(t, &S{Foo: "bar"}, tt.M2)
		})
	})
}
