package configure

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

type I interface {
	action() string
}

type iImpl struct {
	name string
}

func (i *iImpl) action() string {
	return i.name
}

func (i *iImpl) Naming() string {
	return i.name
}

func TestComponentTagExpression(t *testing.T) {
	var config = []byte(`
client: client1
`)
	t.Run("NormalExpression", func(t *testing.T) {
		type T struct {
			I I `wire:"${client}"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.SetConfigLoader(loader.NewRawLoader(config)),
			app.SetComponents(t2,
				&iImpl{name: "defaultClient"},
				&iImpl{name: "client1"},
			),
		)
		assert.NotNil(t, t2.I)
		assert.Equal(t, "client1", t2.I.action())
	})
	t.Run("NormalExpressionWithDefault", func(t *testing.T) {
		type T struct {
			I  I `wire:"${client:defaultClient}"`
			I2 I `wire:"${client2:defaultClient}"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.SetConfigLoader(loader.NewRawLoader(config)),
			app.SetComponents(t2,
				&iImpl{name: "defaultClient"},
				&iImpl{name: "client1"},
			),
		)
		assert.NotNil(t, t2.I)
		assert.Equal(t, "client1", t2.I.action())
		assert.Equal(t, "defaultClient", t2.I2.action())
	})
}
