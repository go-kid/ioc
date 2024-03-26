package configure

import (
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/factory"
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

func TestValueTagExpression(t *testing.T) {
	var config = []byte(`
test:
  host: https://api.dev.go-kid.org
  port:
    - 8080
    - 9090
  parameters:
    header: 
      - X-Request-Id
      - X-Cross-Origin
      - X-Allowed-Method
    aes:
      key: 123
      iv: abc
`)
	t.Run("NormalExpression", func(t *testing.T) {
		type T struct {
			Host       string         `value:"${test.host}"`
			Port       []int          `value:"${test.port}"`
			Parameters map[string]any `value:"${test.parameters}"`
			Headers    []string       `value:"${test.parameters.header}"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.LogTrace,
			app.SetConfigLoader(loader.NewRawLoader(config)),
			app.SetComponents(t2),
		)
		assert.Equal(t, "https://api.dev.go-kid.org", t2.Host)
		assert.Equal(t, []int{8080, 9090}, t2.Port)
		assert.Equal(t, map[string]any{
			"aes": map[string]any{
				"iv":  "abc",
				"key": float64(123),
			},
			"header": []any{"X-Request-Id", "X-Cross-Origin", "X-Allowed-Method"},
		}, t2.Parameters)
		assert.Equal(t, []string{"X-Request-Id", "X-Cross-Origin", "X-Allowed-Method"}, t2.Headers)
	})
	t.Run("NormalExpressionWithDefault", func(t *testing.T) {
		type T struct {
			Host       string         `value:"${test.host2:https://api.go-kid.org}"`
			Port       []int          `value:"${test.port2:[8888,9999]}"`
			PortS      []string       `value:"${test.port2:[:8888,:9999]}"`
			Parameters map[string]any `value:"${test.parameters2:map[a:b]}"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.LogTrace,
			app.SetConfigLoader(loader.NewRawLoader(config)),
			app.SetComponents(t2),
		)
		assert.Equal(t, "https://api.go-kid.org", t2.Host)
		assert.Equal(t, []int{8888, 9999}, t2.Port)
		assert.Equal(t, []string{":8888", ":9999"}, t2.PortS)
		assert.NotNil(t, t2.Parameters)
		assert.Equal(t, map[string]any{
			"a": "b",
		}, t2.Parameters)
	})
	t.Run("MultipleExpression", func(t *testing.T) {
		type T struct {
			Host string `value:"https://${subdomain:api}.${domain:go-kid}.${suffix:org}"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.LogTrace,
			app.SetConfigLoader(loader.NewRawLoader(config)),
			app.SetComponents(t2),
		)
		fmt.Println(t2.Host)
	})
	t.Run("DefaultZeroValue", func(t *testing.T) {
		t.Run("Required", func(t *testing.T) {
			type T struct {
				S string `value:"${t:}${t2:}${t3:}"`
			}
			t2 := &T{}
			err := app.NewApp(
				app.SetRegistry(factory.NewRegistry()),
				app.SetConfigLoader(loader.NewRawLoader(config)),
				app.SetComponents(t2)).Run()
			assert.Error(t, err)
		})
		t.Run("Optional", func(t *testing.T) {
			type T struct {
				S string  `value:"${t:}${t2:}${t3:},required=false"`
				B bool    `value:"${t:},required=false"`
				F float64 `value:"${t:},required=false"`
				I int     `value:"${t:},required=false"`
			}
			t2 := &T{}
			ioc.RunTest(t,
				app.LogTrace,
				app.SetConfigLoader(loader.NewRawLoader(config)),
				app.SetComponents(t2),
			)
			fmt.Println(t2)
		})
	})
}

func TestValueTagValidate(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		type C struct {
			S string `json:"s" validate:"eq=abc"`
		}
		type T struct {
			C *C     `value:"{\"s\":\"abc\"},validate"`
			S string `value:"abc,validate=eq=abc"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.LogTrace,
			app.SetComponents(t2),
		)
	})
	t.Run("Multi-Vars", func(t *testing.T) {
		type T struct {
			S string `value:"123,validate=eq=123 number"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.LogTrace,
			app.SetComponents(t2),
		)
	})
	t.Run("ValidateFailed", func(t *testing.T) {
		t.Run("Struct", func(t *testing.T) {
			type C struct {
				S string `json:"s" validate:"eq=abcd"`
			}
			type T struct {
				C *C `value:"{\"s\":\"abc\"},validate"`
			}
			t2 := &T{}
			err := app.NewApp(
				app.SetRegistry(factory.NewRegistry()),
				app.SetComponents(t2)).Run()
			assert.Error(t, err)
		})
		t.Run("Var", func(t *testing.T) {
			type T struct {
				S string `value:"abc,validate=eq=abcd"`
			}
			t2 := &T{}
			err := app.NewApp(
				app.SetRegistry(factory.NewRegistry()),
				app.SetComponents(t2)).Run()
			assert.Error(t, err)
		})
		t.Run("Multi-Vars", func(t *testing.T) {
			type T struct {
				S string `value:"abc,validate=eq=abc number=true"`
			}
			t2 := &T{}
			err := app.NewApp(
				app.SetRegistry(factory.NewRegistry()),
				app.SetComponents(t2)).Run()
			assert.Error(t, err)
		})
	})
}
