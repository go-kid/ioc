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
	InitHandler func() error
}

func (i *InitializeComponent) Init() error {
	return i.InitHandler()
}

type CloserComponent struct {
	StateComponent
}

func (c *CloserComponent) Close() error {
	c.State = 2
	return nil
}

type RunnerComponent struct {
	StateComponent
	order int
	run   func() error
}

func (r *RunnerComponent) Order() int {
	return r.order
}

func (r *RunnerComponent) Run() error {
	return r.run()
}

func TestComponentLifecycle(t *testing.T) {
	t.Run("TestInitialize", func(t *testing.T) {
		type ChildComp struct {
			InitializeComponent
		}
		type ParentComp struct {
			InitializeComponent
			Child *ChildComp `wire:"Child"`
		}
		var (
			p = &ParentComp{
				InitializeComponent: InitializeComponent{
					StateComponent: StateComponent{
						NamingComponent: NamingComponent{"Parent"},
					},
				}}
			c = &ChildComp{
				InitializeComponent: InitializeComponent{
					StateComponent: StateComponent{
						NamingComponent: NamingComponent{"Child"},
					},
				}}
		)
		p.InitHandler = func() error {
			p.State = 1 + p.Child.State
			return nil
		}
		c.InitHandler = func() error {
			c.State++
			return nil
		}
		ioc.RunTest(t, app.SetComponents(p, c))
		assert.Equal(t, 2, p.State)
		assert.Equal(t, 1, p.Child.State)
		assert.Equal(t, 1, c.State)
	})
	t.Run("TestRunner", func(t *testing.T) {
		var runners []any
		for i := 0; i < 10; i++ {
			runner := &RunnerComponent{
				StateComponent: StateComponent{NamingComponent: NamingComponent{"runner_" + strconv.Itoa(i)}},
				order:          1,
			}
			runner.run = func() error {
				runner.State++
				return nil
			}
			runners = append(runners, runner)
		}
		ioc.RunTest(t, app.SetComponents(runners...))
		for _, runner := range runners {
			assert.Equal(t, 1, runner.(*RunnerComponent).State)
		}
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
		app2 := ioc.RunTest(t, app.LogTrace, app.SetComponents(comps...))
		app2.Close()
		for _, comp := range comps {
			assert.Equal(t, 2, comp.(*CloserComponent).State)
		}
	})
}
