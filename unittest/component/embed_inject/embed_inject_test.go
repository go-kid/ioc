package embed_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ParentComponent struct {
	C  *Component   `wire:"TestComponent"`
	Cs []*Component `wire:""`
}

type ChildComponent1 struct {
	ParentComponent
}

type ChildComponent11 struct {
	ChildComponent1
}

type Component struct {
	Name string
}

func (c *Component) Naming() string {
	return c.Name
}

func TestEmbedComponent(t *testing.T) {
	type tComp struct {
		ChildComponent11
	}
	var tApp = &tComp{}
	ioc.RunTest(t,
		app.LogTrace,
		app.SetComponents(
			&Component{"TestComponent"},
			&Component{"TestComponent2"},
			&Component{},
			tApp,
		))
	assert.Equal(t, "TestComponent", tApp.C.Name)
	assert.Equal(t, 3, len(tApp.Cs))
}
