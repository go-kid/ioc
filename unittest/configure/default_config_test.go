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

func TestConfig(t *testing.T) {
	t.Run("TestYaml", func(t *testing.T) {
		var tApp = &configApp{}
		var _tConfig = `
a:
 b: 123
 c: [ 1,2,3,4 ]
 d:
   d1: "abc"
   d2: 123`
		ioc.RunTest(t, app.SetComponents(tApp),
			app.SetConfig(_tConfig),
			app.SetConfigLoader(loader.NewRawLoader()))
		assert.Equal(t, 123, tApp.A.B)
		assert.Equal(t, []int{1, 2, 3, 4}, tApp.A.C)
		assert.Equal(t, "abc", tApp.D.D1)
		assert.Equal(t, 123, tApp.D.D2)
		assert.Equal(t, 123, tApp.A2.B)
		assert.Equal(t, []int{1, 2, 3, 4}, tApp.A2.C)
	})
	t.Run("TestJson", func(t *testing.T) {
		var tApp = &configApp{}
		var _tConfig = `{"a": {"b": 123, "c": [1,2,3,4], "d": {"d1": "abc", "d2": 123}}}`
		ioc.RunTest(t, app.SetComponents(tApp),
			app.SetConfig(_tConfig),
			app.SetConfigLoader(loader.NewRawLoader()),
			app.SetConfigBinder(binder.NewViperBinder("json")))
		assert.Equal(t, 123, tApp.A.B)
		assert.Equal(t, []int{1, 2, 3, 4}, tApp.A.C)
		assert.Equal(t, "abc", tApp.D.D1)
		assert.Equal(t, 123, tApp.D.D2)
		assert.Equal(t, 123, tApp.A2.B)
		assert.Equal(t, []int{1, 2, 3, 4}, tApp.A2.C)
	})
	t.Run("TestCompare", func(t *testing.T) {
		var tApp = &configApp{}
		var cfg1 = `
a:
 b: 123
 c: [ 1,2,3,4 ]
 d:
   d1: "foo"
   d2: 123`
		var cfg2 = `
a:
 b: 321
 c: [ 1,2,3,4 ]
 d:
   d1: "bar"
   d2: 123`
		iocApp := ioc.RunTest(t, app.SetComponents(tApp),
			app.SetConfig(cfg1),
			app.SetConfigLoader(loader.NewRawLoader()))

		same, err := iocApp.CompareWith([]byte(cfg2), "a.b")
		assert.NoError(t, err)
		assert.False(t, same)

		same, err = iocApp.CompareWith([]byte(cfg2), "a.c")
		assert.NoError(t, err)
		assert.True(t, same)

		same, err = iocApp.CompareWith([]byte(cfg2), "a.d")
		assert.NoError(t, err)
		assert.False(t, same)

		same, err = iocApp.CompareWith([]byte(cfg2), "a.d.d1")
		assert.NoError(t, err)
		assert.False(t, same)

		same, err = iocApp.CompareWith([]byte(cfg2), "a.d.d2")
		assert.NoError(t, err)
		assert.True(t, same)
	})
}
