package ioc

import (
	"context"
	"testing"

	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type runDependency struct{}

type runComponent struct {
	Dependency *runDependency `wire:""`
}

type runContextKey struct{}

type contextAwareComponent struct {
	value any
}

func (c *contextAwareComponent) Init(ctx context.Context) error {
	c.value = ctx.Value(runContextKey{})
	return nil
}

func TestRegisterIsConsumedByNextRun(t *testing.T) {
	registerHandlers = nil
	t.Cleanup(func() { registerHandlers = nil })

	component := &runComponent{}
	dependency := &runDependency{}
	Register(component, dependency)
	require.NotEmpty(t, registerHandlers)

	a, err := Run(app.SkipRunners())
	require.NoError(t, err)
	t.Cleanup(a.Close)
	assert.Same(t, dependency, component.Dependency)
	assert.Empty(t, registerHandlers)
}

func TestRunWithContextPropagatesToInitialization(t *testing.T) {
	component := &contextAwareComponent{}
	ctx := context.WithValue(context.Background(), runContextKey{}, "context-value")

	a, err := RunWithContext(ctx, app.SkipRunners(), app.SetComponents(component))
	require.NoError(t, err)
	t.Cleanup(a.Close)
	assert.Equal(t, "context-value", component.value)
}
