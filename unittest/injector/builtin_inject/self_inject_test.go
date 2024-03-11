package builtin_inject

import (
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ISelfInject interface {
	action()
}

type implSelfInject struct {
}

func (i *implSelfInject) action() {
}

func TestSelfInject(t *testing.T) {
	t.Run("SelfInjectByInterface", func(t *testing.T) {
		type T struct {
			implSelfInject
			InterfaceSelf ISelfInject `wire:""`
		}
		st := &T{}
		err := RunTest(app.SetComponents(st))
		assert.Error(t, err)
	})
	t.Run("SelfInjectByInterfaceSlice", func(t *testing.T) {
		type T struct {
			implSelfInject
			SliceSelf []ISelfInject `wire:""`
		}
		st := &T{}
		err := RunTest(app.SetComponents(st))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(st.SliceSelf))
	})
	t.Run("SelfInjectByPtr", func(t *testing.T) {
		type T struct {
			implSelfInject
			PointerSelf *T `wire:""`
		}
		st := &T{}
		err := RunTest(app.SetComponents(st))
		assert.Error(t, err)
	})
	t.Run("SelfInjectByPtrSlice", func(t *testing.T) {
		type T struct {
			implSelfInject
			SliceSelf []*T `wire:""`
		}
		st := &T{}
		err := RunTest(app.SetComponents(st))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(st.SliceSelf))
	})
}

func RunTest(ops ...app.SettingOption) error {
	s := app.NewApp(append([]app.SettingOption{app.SetRegistry(registry.NewRegistry())}, ops...)...)
	return s.Run()
}
