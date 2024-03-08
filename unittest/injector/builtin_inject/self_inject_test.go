package builtin_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
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
	t.Run("", func(t *testing.T) {
		st := &struct {
			implSelfInject
			InterfaceSelf ISelfInject `wire:""`
		}{}
		ioc.RunTest(t, app.SetComponents(st))
	})
	t.Run("", func(t *testing.T) {
		st := &struct {
			implSelfInject
			SliceSelf []ISelfInject `wire:""`
		}{}
		ioc.RunTest(t, app.SetComponents(st))
	})
	t.Run("", func(t *testing.T) {
		type slfT struct {
			implSelfInject
			PointerSelf *slfT `wire:""`
		}
		st := &slfT{}
		ioc.RunTest(t, app.SetComponents(st))
	})
}
