package special_inject_condition

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

type IdComp interface {
	Id() int
}

type idComp struct {
	id int
}

func (i *idComp) Id() int {
	return i.id
}

type primary struct {
}

func (p *primary) Primary() {
}

type namingComponent struct {
	name string
}

func (c *namingComponent) Naming() string {
	return c.name
}

type SimpleComponent struct {
	idComp
	namingComponent
}

type PrimaryComponent struct {
	idComp
	namingComponent
	primary
}

func TestWirePrimary(t *testing.T) {
	var data = []any{
		&PrimaryComponent{idComp: idComp{id: 0}, namingComponent: namingComponent{name: "naming"}},
		&SimpleComponent{idComp: idComp{id: 1}, namingComponent: namingComponent{name: "simple1"}},
		&SimpleComponent{idComp: idComp{id: 2}, namingComponent: namingComponent{name: "simple2"}},
		&SimpleComponent{idComp: idComp{id: 3}, namingComponent: namingComponent{name: "simple3"}},
	}
	t.Run("PrimarySingleInject", func(t *testing.T) {
		type T struct {
			Comp IdComp `wire:""`
		}
		for i := 0; i < 100; i++ {
			var tt = &T{}
			ioc.RunTest(t, app.LogError, app.SetComponents(data...), app.SetComponents(tt))
			assert.Equal(t, 0, tt.Comp.Id())
		}
	})
	t.Run("PrimarySliceInject", func(t *testing.T) {
		type T struct {
			Comp []IdComp `wire:""`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(data...), app.SetComponents(tt))
		assert.Equal(t, len(data), len(tt.Comp))
	})
}
