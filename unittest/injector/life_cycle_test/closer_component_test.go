package life_cycle_test

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

type Component struct {
	Name string
}

type NamingComponent Component

func (c *NamingComponent) Naming() string {
	return c.Name
}

type StateComponent struct {
	NamingComponent
	State int
}

type InitializeComponent struct {
	StateComponent
	Child *InitializeComponent `wire:"Child"`
}

func (i *InitializeComponent) Init() error {
	i.State = 1 + i.Child.State
	return nil
}

type CloserComponent struct {
	StateComponent
}

func (c *CloserComponent) Close() error {
	c.State = 2
	return nil
}

func TestComponentLifecycle(t *testing.T) {
	t.Run("TestInitialize", func(t *testing.T) {
		var (
			p = &InitializeComponent{
				StateComponent: StateComponent{
					NamingComponent: NamingComponent{"Parent"},
				},
			}
			c = &InitializeComponent{
				StateComponent: StateComponent{
					NamingComponent: NamingComponent{"Child"},
				},
			}
		)

		ioc.RunTest(t, app.SetComponents(p, c))
		assert.Equal(t, 2, p.State)
		assert.Equal(t, 1, p.Child.State)
		assert.Equal(t, 1, c.State)
		assert.Equal(t, 1, c.Child.State)
	})
	t.Run("TestClose", func(t *testing.T) {
		var comps []any
		for i := 0; i < 10; i++ {
			comps = append(comps, &CloserComponent{StateComponent{
				NamingComponent: NamingComponent{
					Name: strconv.Itoa(i),
				},
			}})
		}
		app2 := ioc.RunTest(t, app.SetComponents(comps...))
		app2.Close()
		for _, comp := range comps {
			assert.Equal(t, 2, comp.(*CloserComponent).State)
		}
	})
}
