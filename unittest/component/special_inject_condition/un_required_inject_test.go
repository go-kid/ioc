package special_inject_condition

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"testing"
)

func TestUnRequiredInject(t *testing.T) {
	t.Run("DefaultRequiredSingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:""`
		}
		var tt = &T{}
		ioc.RunErrorTest(t, app.SetComponents(tt))
	})
	t.Run("RequiredArgSingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:",required=true"`
		}
		var tt = &T{}
		ioc.RunErrorTest(t, app.SetComponents(tt))
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
		ioc.RunErrorTest(t, app.SetComponents(tt))
	})
	t.Run("RequiredArgSliceInject", func(t *testing.T) {
		type T struct {
			Comp []ITest `wire:",required=true"`
		}
		var tt = &T{}
		ioc.RunErrorTest(t, app.SetComponents(tt))
	})
	t.Run("UnRequiredArgSliceInject", func(t *testing.T) {
		type T struct {
			Comp []ITest `wire:",required=false"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(tt))
	})
}
