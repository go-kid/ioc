package configure

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/stretchr/testify/assert"
	"testing"
)

type configA struct {
	B int   `yaml:"b"`
	C []int `yaml:"c"`
}

type configD struct {
	D1 string `yaml:"d1"`
	D2 int    `yaml:"d2"`
}

func (c *configD) Prefix() string {
	return "a.d"
}

type configApp struct {
	A  *configA `prop:"a"`
	D  *configD
	A2 configA `prop:"a"`
}

func TestBinder(t *testing.T) {
	t.Run("TestYaml", func(t *testing.T) {
		var tApp = &configApp{}
		var _tConfig = []byte(`
a:
 b: 123
 c: [ 1,2,3,4 ]
 d:
   d1: "abc"
   d2: 123`)
		ioc.RunTest(t, app.SetComponents(tApp),
			app.SetConfigLoader(loader.NewRawLoader(_tConfig)))
		assert.Equal(t, 123, tApp.A.B)
		assert.Equal(t, []int{1, 2, 3, 4}, tApp.A.C)
		assert.Equal(t, "abc", tApp.D.D1)
		assert.Equal(t, 123, tApp.D.D2)
		assert.Equal(t, 123, tApp.A2.B)
		assert.Equal(t, []int{1, 2, 3, 4}, tApp.A2.C)
	})
	t.Run("TestJson", func(t *testing.T) {
		var tApp = &configApp{}
		var _tConfig = []byte(`{"a": {"b": 123, "c": [1,2,3,4], "d": {"d1": "abc", "d2": 123}}}`)
		ioc.RunTest(t, app.SetComponents(tApp),
			app.SetConfigLoader(loader.NewRawLoader(_tConfig)),
			app.SetConfigBinder(binder.NewViperBinder("json")))
		assert.Equal(t, 123, tApp.A.B)
		assert.Equal(t, []int{1, 2, 3, 4}, tApp.A.C)
		assert.Equal(t, "abc", tApp.D.D1)
		assert.Equal(t, 123, tApp.D.D2)
		assert.Equal(t, 123, tApp.A2.B)
		assert.Equal(t, []int{1, 2, 3, 4}, tApp.A2.C)
	})
	t.Run("TestGet", func(t *testing.T) {
		var cfg1 = []byte(`
a:
 b: 123
 c: [ 1,2,3,4 ]
 d:
   d1: "foo"
   d2: 123`)
		iocApp := ioc.RunTest(t,
			app.SetConfigLoader(loader.NewRawLoader(cfg1)))

		val := iocApp.Get("a.b")
		assert.Equal(t, 123, val)

		val = iocApp.Get("a.c")
		assert.Equal(t, []any{1, 2, 3, 4}, val)

		val = iocApp.Get("a.d.d1")
		assert.Equal(t, "foo", val)

		val = iocApp.Get("a.d.d2")
		assert.Equal(t, 123, val)
	})
	t.Run("TestSet", func(t *testing.T) {
		iocApp := ioc.RunTest(t)
		iocApp.Set("a.b", "123")
		iocApp.Set("a.c", 123)
		iocApp.Set("b.a", []string{"foo", "bar"})
		assert.Equal(t, "123", iocApp.Get("a.b"))
		assert.Equal(t, 123, iocApp.Get("a.c"))
		assert.Equal(t, []string{"foo", "bar"}, iocApp.Get("b.a"))
	})
}

func TestConfigureTagExpression(t *testing.T) {
	var config = []byte(`
env: dev
test:
  dev:
    host: https://api.dev.go-kid.org
  local:
    host: http://localhost:8080
`)
	t.Run("NormalExpression", func(t *testing.T) {
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
	t.Run("NormalExpressionWithDefault", func(t *testing.T) {
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
