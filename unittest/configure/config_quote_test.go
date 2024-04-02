package configure

import (
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigureConfigQuote(t *testing.T) {
	var config = []byte(`
env: dev
test:
  dev:
    host: https://api.dev.go-kid.org
  local:
    host: http://localhost:8080
`)
	t.Run("NormalConfigQuote", func(t *testing.T) {
		type T struct {
			Host string `prop:"test.${env}.host"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.LogTrace,
			app.SetConfigLoader(loader.NewRawLoader(config)),
			app.SetComponents(t2),
		)
		assert.Equal(t, "https://api.dev.go-kid.org", t2.Host)
	})
	t.Run("NormalConfigQuoteWithDefault", func(t *testing.T) {
		type T struct {
			Host  string `prop:"test.${env:local}.host"`
			Host2 string `prop:"test.${env2:local}.host"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.LogTrace,
			app.SetConfigLoader(loader.NewRawLoader(config)),
			app.SetComponents(t2),
		)
		assert.Equal(t, "https://api.dev.go-kid.org", t2.Host)
		assert.Equal(t, "http://localhost:8080", t2.Host2)
	})
}

func TestValueConfigQuote(t *testing.T) {
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
	t.Run("NormalConfigQuote", func(t *testing.T) {
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
	t.Run("NormalConfigQuoteWithDefault", func(t *testing.T) {
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
	t.Run("MultipleConfigQuote", func(t *testing.T) {
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
			ioc.RunErrorTest(t, app.SetConfigLoader(loader.NewRawLoader(config)),
				app.SetComponents(t2))
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
