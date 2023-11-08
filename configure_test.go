package ioc

import (
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/stretchr/testify/assert"
	"testing"
)

const expressionYaml = `
App:
  Name: expression
  Env: dev
  Port: ":8080"
Http:
  address: api.{{.app.env}}.xxx.com
Host: "{{.app.name}}.{{.app.name}}-{{.app.env}}.svc.cluster{{.app.port}}"
`

type expressionTestApp struct {
	C *expressionConfig `prop:""`
}

type expressionConfig struct {
	App struct {
		Name string `yaml:"Name"`
		Env  string `yaml:"Env"`
		Port string `yaml:"Port"`
	} `yaml:"App"`
	Http struct {
		Address string `yaml:"Address"`
	} `yaml:"Http"`
	Host string `yaml:"Host"`
}

func TestExpressionBinder(t *testing.T) {
	var a = &expressionTestApp{}
	RunTest(t,
		app.SetConfigLoader(loader.NewRawLoader()),
		app.SetConfigBinder(binder.NewExpressionBinder("yaml")),
		app.SetConfig(expressionYaml),
		app.SetComponents(a))
	assert.Equal(t, "api.dev.xxx.com", a.C.Http.Address)
	assert.Equal(t, "expression.expression-dev.svc.cluster:8080", a.C.Host)
}
