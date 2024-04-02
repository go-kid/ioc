package special_inject_condition

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
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
		ioc.RunErrorTest(t, app.SetComponents(st), app.LogTrace)
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
		ioc.RunErrorTest(t, app.SetComponents(st), app.LogTrace)
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
		ioc.RunErrorTest(t, app.SetComponents(st))
	})
	t.Run("SelfInjectByPtr", func(t *testing.T) {
		type T struct {
			implSelfInject
			PointerSelf *T `wire:""`
		}
		st := &T{}
		ioc.RunErrorTest(t, app.SetComponents(st))
	})
	t.Run("SelfInjectByPtrSlice", func(t *testing.T) {
		type T struct {
			implSelfInject
			SliceSelf []*T `wire:""`
		}
		st := &T{}
		ioc.RunErrorTest(t, app.SetComponents(st))
	})
}
