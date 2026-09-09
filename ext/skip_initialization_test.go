package ext_test

import (
	"testing"

	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type skipDependency struct{}

type skipComponent struct {
	Dependency  *skipDependency `wire:""`
	Initialized bool
}

func (c *skipComponent) Init() error {
	c.Initialized = true
	return nil
}

type skipRunner struct{ Ran bool }

func (r *skipRunner) Run() error {
	r.Ran = true
	return nil
}

func TestSkipComponentInitialization(t *testing.T) {
	dependency := &skipDependency{}
	component := &skipComponent{}
	runner := &skipRunner{}

	a, err := ioc.Run(
		ext.SkipComponentInitialization(),
		app.SetComponents(dependency, component, runner),
	)
	require.NoError(t, err)
	t.Cleanup(a.Close)

	assert.Same(t, dependency, component.Dependency)
	assert.False(t, component.Initialized)
	assert.False(t, runner.Ran)
}
