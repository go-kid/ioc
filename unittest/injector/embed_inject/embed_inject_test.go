package embed_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ParentComponent struct {
	C *Component `wire:""`
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

func TestEmbedComponent(t *testing.T) {
	var tApp = &struct {
		ChildComponent11
	}{}
	ioc.RunTest(t, app.SetComponents(
		&Component{"TestComponent"},
		tApp,
	))
	assert.Equal(t, "TestComponent", tApp.C.Name)
}
