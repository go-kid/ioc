package ioc

import (
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

func TestConfig(t *testing.T) {
	var app = &struct {
		A *configA `prop:"a"`
		D *configD
	}{}
	RunTest(t, "config.yaml", SetComponents(app))
	assert.Equal(t, 123, app.A.B)
	assert.Equal(t, "abc", app.D.D1)
}
