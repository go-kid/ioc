package configure

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValueTag(t *testing.T) {
	t.Run("TestSingleValue", func(t *testing.T) {
		type T struct {
			A string   `value:"foo"`
			S []string `value:"[hello world foo bar]"`
			B bool     `value:"true"`
			I int      `value:"123"`
			F float64  `value:"123.321"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(tt))
		assert.Equal(t, "foo", tt.A)
		assert.Equal(t, []string{"hello", "world", "foo", "bar"}, tt.S)
		assert.True(t, tt.B)
		assert.Equal(t, 123, tt.I)
		assert.Equal(t, 123.321, tt.F)
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
    header: [X-Request-Id X-Cross-Origin X-Allowed-Method]
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
	})
	t.Run("NormalExpressionWithDefault", func(t *testing.T) {
		type T struct {
			Host string `value:"${test.host2:https://api.go-kid.org}"`
			Port []int  `value:"${test.port2:[8888 9999]}"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.LogTrace,
			app.SetConfigLoader(loader.NewRawLoader(config)),
			app.SetComponents(t2),
		)
		assert.Equal(t, "https://api.go-kid.org", t2.Host)
		assert.Equal(t, []int{8888, 9999}, t2.Port)
	})
}
