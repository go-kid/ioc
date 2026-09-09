package refactor_test

import (
	"testing"

	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type conditionalDependency struct{}

func (*conditionalDependency) Naming() string { return "conditionalDependency" }

type conditionalProbe struct {
	hasDependency bool
	configValue   any
	initialized   bool
}

func (p *conditionalProbe) Condition(ctx definition.ConditionContext) bool {
	p.hasDependency = ctx.HasComponent("conditionalDependency")
	p.configValue = ctx.GetConfig("feature.enabled")
	return p.hasDependency && p.configValue == true
}

func (p *conditionalProbe) Init() error {
	p.initialized = true
	return nil
}

func TestConditionalContextExposesRegistryAndConfiguration(t *testing.T) {
	probe := &conditionalProbe{}
	a := app.NewApp()
	err := a.Run(
		app.SetConfigLoader(loader.NewRawLoader([]byte("feature:\n  enabled: true\n"))),
		app.SetComponents(&conditionalDependency{}, probe),
	)
	require.NoError(t, err)
	t.Cleanup(a.Close)

	assert.True(t, probe.hasDependency)
	assert.Equal(t, true, probe.configValue)
	assert.True(t, probe.initialized)
}
