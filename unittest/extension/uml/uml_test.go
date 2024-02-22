package uml

import (
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/extension/uml"
	"github.com/go-kid/ioc/unittest/extension/uml/config"
	"github.com/go-kid/ioc/unittest/extension/uml/def"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type NodeA struct {
	Cfg      *config.Config
	NodeOne  def.I   `wire:""`
	NodeList []def.L `wire:""`
}

type NodeB struct {
	NodeA *NodeA `wire:""`
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
