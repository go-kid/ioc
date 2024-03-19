package special_inject_condition

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnRequiredInject(t *testing.T) {
	t.Run("DefaultRequiredSingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:""`
		}
		var tt = &T{}
		newApp := app.NewApp(app.SetRegistry(registry.NewRegistry()), app.SetComponents(tt))
		err := newApp.Run()
		assert.Error(t, err)
	})
	t.Run("RequiredArgSingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:",required=true"`
		}
		var tt = &T{}
		newApp := app.NewApp(app.SetRegistry(registry.NewRegistry()), app.SetComponents(tt))
		err := newApp.Run()
		assert.Error(t, err)
	})
	t.Run("UnRequiredArgSingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:",required=false"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(tt))
	})

	t.Run("DefaultRequiredSliceInject", func(t *testing.T) {
		type T struct {
			Comp []ITest `wire:""`
		}
		var tt = &T{}
		newApp := app.NewApp(app.SetRegistry(registry.NewRegistry()), app.SetComponents(tt))
		err := newApp.Run()
		assert.Error(t, err)
	})
	t.Run("RequiredArgSliceInject", func(t *testing.T) {
		type T struct {
			Comp []ITest `wire:",required=true"`
		}
		var tt = &T{}
		newApp := app.NewApp(app.SetRegistry(registry.NewRegistry()), app.SetComponents(tt))
		err := newApp.Run()
		assert.Error(t, err)
	})
	t.Run("UnRequiredArgSliceInject", func(t *testing.T) {
		type T struct {
			Comp []ITest `wire:",required=false"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(tt))
	})
}
