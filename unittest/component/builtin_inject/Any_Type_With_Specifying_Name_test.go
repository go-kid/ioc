package builtin_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/defination"
	"github.com/stretchr/testify/assert"
	"testing"
)

type SpecifyNameComponent struct {
	Component
}

func (s *SpecifyNameComponent) Naming() string {
	return s.Name
}

func Test_Any_Type_With_Specifying_Name(t *testing.T) {
	t.Run("Pointer_Type", func(t *testing.T) {
		var tApp = &struct {
			T *SpecifyNameComponent `wire:"TestComponent_1"`
		}{}
		ioc.RunTest(t, app.SetComponents(
			&SpecifyNameComponent{Component{"TestComponent_1"}},
			&SpecifyNameComponent{Component{"TestComponent_2"}},
			tApp,
		))
		assert.Equal(t, "TestComponent_1", tApp.T.Name)
	})
	t.Run("Interface_Type", func(t *testing.T) {
		var tApp = &struct {
			T defination.NamingComponent `wire:"TestComponent_1"`
		}{}
		ioc.RunTest(t, app.SetComponents(
			&SpecifyNameComponent{Component{"TestComponent_1"}},
			&SpecifyNameComponent{Component{"TestComponent_2"}},
			tApp,
		))
		assert.Equal(t, "TestComponent_1", tApp.T.(*SpecifyNameComponent).Name)
	})
	t.Run("Any_Type", func(t *testing.T) {
		var tApp = &struct {
			T any `wire:"TestComponent_1"`
		}{}
		ioc.RunTest(t, app.SetComponents(
			&SpecifyNameComponent{Component{"TestComponent_1"}},
			&SpecifyNameComponent{Component{"TestComponent_2"}},
			tApp,
		))
		assert.Equal(t, "TestComponent_1", tApp.T.(*SpecifyNameComponent).Name)
	})
}
