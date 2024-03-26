package special_inject_condition

import (
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/factory"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Component struct {
	Name string
}

type SpecifyNameComponent struct {
	Component
}

func (s *SpecifyNameComponent) Naming() string {
	return s.Name
}

type ISelfInject interface {
	action()
}

type implSelfInject struct {
	SpecifyNameComponent
}

func (i *implSelfInject) action() {
}

func TestSelfInject(t *testing.T) {
	t.Run("SelfInjectByInterface", func(t *testing.T) {
		type T struct {
			implSelfInject
			InterfaceSelf ISelfInject `wire:""`
		}
		type T2 struct {
			T
		}
		type T3 struct {
			T2
		}
		st := &T3{}
		err := RunTest(app.SetComponents(st), app.LogTrace)
		assert.Error(t, err)
	})
	t.Run("SelfInjectByNamingInterface", func(t *testing.T) {
		type T struct {
			implSelfInject
			InterfaceSelf ISelfInject `wire:"t1"`
		}
		type T2 struct {
			T
		}
		type T3 struct {
			T2
		}
		st := &T3{T2: T2{T: T{implSelfInject: implSelfInject{SpecifyNameComponent{Component{Name: "t1"}}}}}}
		err := RunTest(app.SetComponents(st), app.LogTrace)
		assert.Error(t, err)
	})
	t.Run("SelfInjectByInterfaceSlice", func(t *testing.T) {
		type T struct {
			implSelfInject
			SliceSelf []ISelfInject `wire:""`
		}
		type T2 struct {
			T
		}
		type T3 struct {
			T2
		}
		st := &T3{}
		err := RunTest(app.SetComponents(st))
		assert.Error(t, err)
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
		assert.Error(t, err)
	})
}

func RunTest(ops ...app.SettingOption) error {
	s := app.NewApp(append([]app.SettingOption{app.SetRegistry(factory.NewRegistry())}, ops...)...)
	return s.Run()
}
