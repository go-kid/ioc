package uml

import (
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/extension/uml"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type Config struct {
	Val1 int
	Val2 string
}

func (c *Config) Prefix() string {
	return "config"
}

type NodeA struct {
	Cfg      *Config
	NodeOne  I   `wire:""`
	NodeList []L `wire:""`
}

type NodeB struct {
	NodeA *NodeA `wire:""`
}

type I interface {
	Do()
}

type L interface {
	LDo() string
}

type Base struct {
	NodeB *NodeB `wire:""`
}

type NodeC struct {
	Base
	NodeC *NodeC `wire:""`
}

func (c *NodeC) Do() {}

func (c *NodeC) LDo() string {
	return ""
}

func TestUML(t *testing.T) {
	dataWriter := &strings.Builder{}
	_, err := uml.Run(
		uml.DebugSetting{
			DisablePackageView:      true,
			DisableConfig:           false,
			DisableConfigDetail:     false,
			DisableDependency:       false,
			DisableDependencyDetail: false,
			DisableUselessClass:     false,
			RawData:                 dataWriter,
			StartServer:             true,
		}, app.SetComponents(
			&NodeA{},
			&NodeB{},
			&NodeC{},
		),
	)
	assert.NoError(t, err)
}
