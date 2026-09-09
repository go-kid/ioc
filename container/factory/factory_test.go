package factory

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/container/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lifecycleComponent struct {
	initialized bool
}

func (c *lifecycleComponent) Init() error {
	c.initialized = true
	return nil
}

type lifecycleProcessor struct {
	beforeCalled  bool
	afterCalled   bool
	destroyCalled bool
	destroyErr    error
}

func (p *lifecycleProcessor) PostProcessBeforeInitialization(component any, _ string) (any, error) {
	p.beforeCalled = true
	return component, nil
}

func (p *lifecycleProcessor) PostProcessAfterInitialization(component any, _ string) (any, error) {
	p.afterCalled = true
	return component, nil
}

func (p *lifecycleProcessor) RequireDestruction(component any) bool {
	_, ok := component.(*lifecycleComponent)
	return ok
}

func (p *lifecycleProcessor) PostProcessBeforeDestruction(any, string) error {
	p.destroyCalled = true
	return p.destroyErr
}

func TestPostProcessorRegistrationDelegateLifecycle(t *testing.T) {
	processor := &lifecycleProcessor{}
	delegate := NewPostProcessorRegistrationDelegate()
	delegate.RegisterComponentPostProcessors(processor, "processor")
	delegate.componentPostProcessors = append(delegate.componentPostProcessors, processor)
	component := &lifecycleComponent{}

	result, err := delegate.InitializeComponentWithContext(context.Background(), "component", component)
	require.NoError(t, err)
	assert.Same(t, component, result)
	assert.True(t, processor.beforeCalled)
	assert.True(t, component.initialized)
	assert.True(t, processor.afterCalled)

	require.NoError(t, delegate.DestroyComponent(component, "component"))
	assert.True(t, processor.destroyCalled)
}

func TestPostProcessorRegistrationDelegateDestructionError(t *testing.T) {
	wantErr := errors.New("destroy failed")
	processor := &lifecycleProcessor{destroyErr: wantErr}
	delegate := NewPostProcessorRegistrationDelegate()
	delegate.RegisterComponentPostProcessors(processor, "processor")
	delegate.componentPostProcessors = append(delegate.componentPostProcessors, processor)

	err := delegate.DestroyComponent(&lifecycleComponent{}, "component")
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
	assert.Contains(t, err.Error(), "PostProcessBeforeDestruction")
}

func TestFactoryContextConditionAndDependencyChain(t *testing.T) {
	registry := support.NewRegistry()
	registry.RegisterSingleton(&lifecycleComponent{})
	configuration := configure.Default()
	configuration.Set("feature.enabled", true)
	f := Default().(*defaultFactory)
	f.SetRegistry(registry)
	f.SetConfigure(configuration)

	condition := f.newConditionContext()
	assert.True(t, condition.HasComponent(registry.GetSingletonNames()[0]))
	assert.False(t, condition.HasComponent("missing"))
	assert.Equal(t, true, condition.GetConfig("feature.enabled"))

	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "value")
	f.SetContext(ctx)
	assert.Equal(t, "value", f.getContext().Value(key{}))

	f.resolveStack = []string{"service", "repository"}
	assert.Equal(t, "dependency resolution failed:\nservice\n  -> repository\n    -> database (not found)", f.formatDependencyChain("database", "not found"))
}

var _ container.DestructionAwareComponentPostProcessor = (*lifecycleProcessor)(nil)
