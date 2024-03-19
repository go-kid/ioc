package builtin_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Interface_Slice(t *testing.T) {
	var tApp = &struct {
		Ts []Interface `wire:""`
	}{}
	ioc.RunTest(t, app.SetComponents(
		&InterfaceImplNamingComponent{SpecifyNameComponent{Component{"InterfaceImplNamingComponent"}}},
		&InterfaceImplComponent{Component{"InterfaceImplComponent"}},
		&SpecifyNameComponent{Component{"SpecifyNameComponent"}},
		&Component{"Component"},
		tApp,
	))
	assert.Equal(t, 2, len(tApp.Ts))
}
