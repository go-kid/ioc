package builtin_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Interface interface {
	SimpleInterface()
}

type InterfaceImplComponent struct {
	Component
}

func (i *InterfaceImplComponent) SimpleInterface() {}

type InterfaceImplNamingComponent struct {
	SpecifyNameComponent
}

func (i *InterfaceImplNamingComponent) SimpleInterface() {}

func Test_Interface_Without_Specifying_Name(t *testing.T) {
	t.Run("Interface_Type", func(t *testing.T) {
		var tApp = &struct {
			T Interface `wire:""`
		}{}
		ioc.RunTest(t, app.SetComponents(
			&InterfaceImplComponent{Component{"TestComponent"}},
			tApp,
		))
		assert.Equal(t, "TestComponent", tApp.T.(*InterfaceImplComponent).Name)
	})
	t.Run("Nameless_Prefer", func(t *testing.T) {
		var tApp = &struct {
			T Interface `wire:""`
		}{}
		ioc.RunTest(t, app.SetComponents(
			&InterfaceImplComponent{Component{"TestComponent_Nameless"}},
			&InterfaceImplNamingComponent{SpecifyNameComponent{Component{"TestComponent_Naming"}}},
			tApp,
		))
		assert.Equal(t, "TestComponent_Nameless", tApp.T.(*InterfaceImplComponent).Name)
	})
	t.Run("Nameless_Prefer_Default", func(t *testing.T) {
		var tApp = &struct {
			T Interface `wire:""`
		}{}
		ioc.RunTest(t, app.SetComponents(
			&InterfaceImplNamingComponent{SpecifyNameComponent{Component{"TestComponent_Naming"}}},
			tApp,
		))
		assert.Equal(t, "TestComponent_Naming", tApp.T.(*InterfaceImplNamingComponent).Name)
	})
}
