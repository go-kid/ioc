// Package ext provides optional, opt-in extensions for the ioc container that
// are not part of the default startup behavior.
package ext

import (
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/container/processors"
)

// skipInitializationOrder makes the processor run last among ordered
// ComponentPostProcessors so that any meaningful before-initialization logic of
// the built-in processors still executes before initialization is skipped.
const skipInitializationOrder = 1 << 30

// SkipInitializationPostProcessor skips the initialization phase of every
// component. By returning nil from PostProcessBeforeInitialization it
// short-circuits Factory's InitializeComponentWithContext before invokeInitMethods
// runs, so neither Init()/Init(ctx) nor AfterPropertiesSet() is called.
//
// Dependency injection (wire/value/prop/prefix/func) is unaffected: it completes
// during the populate phase, which happens before initialization.
type SkipInitializationPostProcessor struct {
	processors.DefaultComponentPostProcessor
}

// NewSkipInitializationPostProcessor builds a SkipInitializationPostProcessor.
func NewSkipInitializationPostProcessor() container.ComponentPostProcessor {
	return &SkipInitializationPostProcessor{}
}

// PostProcessBeforeInitialization returns nil to skip the component's
// initialization methods. The factory keeps and exposes the already-populated
// instance unchanged.
func (p *SkipInitializationPostProcessor) PostProcessBeforeInitialization(component any, componentName string) (any, error) {
	return nil, nil
}

// Order places this processor at the end of the ordered ComponentPostProcessor chain.
func (p *SkipInitializationPostProcessor) Order() int {
	return skipInitializationOrder
}

// SkipComponentInitialization returns a SettingOption that performs dependency
// injection only: every component's Init()/AfterPropertiesSet() is skipped and
// all ApplicationRunner.Run() invocations are skipped as well.
//
// It is intended for local development when real dependencies (databases,
// cluster services, etc.) are unreachable but the wiring/injection capability is
// still needed. Note that since initialization is short-circuited, the
// after-initialization processor chain (e.g. AOP proxying done in
// PostProcessAfterInitialization) is also skipped for every component.
//
// Usage:
//
//	ioc.Run(ext.SkipComponentInitialization(), app.SetComponents(...))
func SkipComponentInitialization() app.SettingOption {
	return app.Options(
		app.SkipRunners(),
		app.SetComponents(NewSkipInitializationPostProcessor()),
	)
}
