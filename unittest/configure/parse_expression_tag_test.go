package configure

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/syslog"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseExpressionTag(t *testing.T) {
	var config = `
env: dev
test:
  dev:
    host: https://api.dev.go-kid.org
  local:
    host: http://localhost:8080
`
	t.Run("NormalExpression", func(t *testing.T) {
		type T struct {
			Host string `prop:"test.${env}.host"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.LogLevel(syslog.LvTrace),
			app.SetConfig(config),
			app.SetConfigLoader(loader.NewRawLoader()),
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
			app.LogLevel(syslog.LvTrace),
			app.SetConfig(config),
			app.SetConfigLoader(loader.NewRawLoader()),
			app.SetComponents(t2),
		)
		assert.Equal(t, "https://api.dev.go-kid.org", t2.Host)
		assert.Equal(t, "http://localhost:8080", t2.Host2)
	})
}
