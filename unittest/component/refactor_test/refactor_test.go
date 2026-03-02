package refactor_test

import (
	"context"
	"sync/atomic"
	"testing"

	ioc "github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/definition"
	"github.com/stretchr/testify/assert"
)

// --- Context lifecycle tests ---

type ctxTracker struct {
	initCtx              context.Context
	afterPropertiesCtx   context.Context
	runCtx               context.Context
	closeCtx             context.Context
}

type ctxInitComponent struct {
	tracker *ctxTracker
}

func (c *ctxInitComponent) Init(ctx context.Context) error {
	c.tracker.initCtx = ctx
	return nil
}

type ctxAfterPropertiesComponent struct {
	tracker *ctxTracker
}

func (c *ctxAfterPropertiesComponent) AfterPropertiesSet(ctx context.Context) error {
	c.tracker.afterPropertiesCtx = ctx
	return nil
}

type ctxRunner struct {
	tracker *ctxTracker
}

func (r *ctxRunner) Run() error { return nil }
func (r *ctxRunner) RunWithContext(ctx context.Context) error {
	r.tracker.runCtx = ctx
	return nil
}

type ctxCloser struct {
	tracker *ctxTracker
}

func (c *ctxCloser) Close() error { return nil }
func (c *ctxCloser) CloseWithContext(ctx context.Context) error {
	c.tracker.closeCtx = ctx
	return nil
}

func TestContextPropagation(t *testing.T) {
	tracker := &ctxTracker{}
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "test-value")

	a := app.NewApp()
	err := a.RunWithContext(ctx,
		app.SetComponents(
			&ctxInitComponent{tracker: tracker},
			&ctxAfterPropertiesComponent{tracker: tracker},
			&ctxRunner{tracker: tracker},
			&ctxCloser{tracker: tracker},
		),
	)
	assert.NoError(t, err)

	assert.Equal(t, "test-value", tracker.initCtx.Value(ctxKey{}))
	assert.Equal(t, "test-value", tracker.afterPropertiesCtx.Value(ctxKey{}))
	assert.Equal(t, "test-value", tracker.runCtx.Value(ctxKey{}))

	a.CloseWithContext(ctx)
	assert.Equal(t, "test-value", tracker.closeCtx.Value(ctxKey{}))
}

// --- Backward compatibility: old interfaces still work ---

type legacyInitComponent struct {
	initCalled bool
}

func (c *legacyInitComponent) Init() error {
	c.initCalled = true
	return nil
}

type legacyAfterPropertiesComponent struct {
	called bool
}

func (c *legacyAfterPropertiesComponent) AfterPropertiesSet() error {
	c.called = true
	return nil
}

type legacyRunner struct {
	called bool
}

func (r *legacyRunner) Run() error {
	r.called = true
	return nil
}

type legacyCloser struct {
	called bool
}

func (c *legacyCloser) Close() error {
	c.called = true
	return nil
}

func TestBackwardCompatibility_OldInterfaces(t *testing.T) {
	initComp := &legacyInitComponent{}
	afterComp := &legacyAfterPropertiesComponent{}
	runner := &legacyRunner{}
	closer := &legacyCloser{}

	a := app.NewApp()
	err := a.Run(
		app.SetComponents(initComp, afterComp, runner, closer),
	)
	assert.NoError(t, err)
	assert.True(t, initComp.initCalled)
	assert.True(t, afterComp.called)
	assert.True(t, runner.called)

	a.Close()
	assert.True(t, closer.called)
}

// --- Scope tests ---

type prototypeService struct {
	id int32
}

var protoCounter int32

func (p *prototypeService) Scope() string { return definition.ScopePrototype }
func (p *prototypeService) Naming() string {
	return "prototypeService"
}

func (p *prototypeService) Init() error {
	p.id = atomic.AddInt32(&protoCounter, 1)
	return nil
}

type singletonService struct {
	id int32
}

func (s *singletonService) Init() error {
	s.id = 42
	return nil
}

func TestScopeDefault_Singleton(t *testing.T) {
	svc := &singletonService{}
	a := app.NewApp()
	err := a.Run(app.SetComponents(svc))
	assert.NoError(t, err)
	assert.Equal(t, int32(42), svc.id)
}

// --- Conditional tests ---

type alwaysSkipped struct {
	initCalled bool
}

func (a *alwaysSkipped) Condition(ctx definition.ConditionContext) bool {
	return false
}

func (a *alwaysSkipped) Init() error {
	a.initCalled = true
	return nil
}

type alwaysIncluded struct {
	initCalled bool
}

func (a *alwaysIncluded) Condition(ctx definition.ConditionContext) bool {
	return true
}

func (a *alwaysIncluded) Init() error {
	a.initCalled = true
	return nil
}

func TestConditional_SkipFalse(t *testing.T) {
	skipped := &alwaysSkipped{}
	included := &alwaysIncluded{}

	a := app.NewApp()
	err := a.Run(app.SetComponents(skipped, included))
	assert.NoError(t, err)
	assert.False(t, skipped.initCalled)
	assert.True(t, included.initCalled)
}

// --- Event tests ---

type testEventListener struct {
	events []definition.ApplicationEvent
}

func (l *testEventListener) OnEvent(event definition.ApplicationEvent) error {
	l.events = append(l.events, event)
	return nil
}

func TestEventMechanism(t *testing.T) {
	listener := &testEventListener{}

	a := app.NewApp()
	err := a.Run(app.SetComponents(listener))
	assert.NoError(t, err)

	var hasStarted bool
	for _, e := range listener.events {
		if _, ok := e.(*definition.ApplicationStartedEvent); ok {
			hasStarted = true
		}
	}
	assert.True(t, hasStarted, "should receive ApplicationStartedEvent")

	a.Close()
	var hasClosing bool
	for _, e := range listener.events {
		if _, ok := e.(*definition.ApplicationClosingEvent); ok {
			hasClosing = true
		}
	}
	assert.True(t, hasClosing, "should receive ApplicationClosingEvent")
}

// --- ioc.Run backward compatibility ---

func TestIocRun_BackwardCompat(t *testing.T) {
	runner := &legacyRunner{}
	a, err := ioc.Run(app.SetComponents(runner))
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.True(t, runner.called)
}

func TestIocRunWithContext(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "hello")
	tracker := &ctxTracker{}

	a, err := ioc.RunWithContext(ctx, app.SetComponents(
		&ctxRunner{tracker: tracker},
	))
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.Equal(t, "hello", tracker.runCtx.Value(ctxKey{}))
}
