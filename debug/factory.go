package debug

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/container/factory"
)

type hookSetter interface {
	SetFactoryHook(hook container.FactoryHook)
}

type contextSetter interface {
	SetContext(ctx context.Context)
}

type DebugFactory struct {
	inner      container.Factory
	controller *Controller
	server     *Server
	collector  *Collector
	hook       *debugHook
}

func NewDebugFactory() *DebugFactory {
	inner := factory.Default()
	ctrl := NewController()
	coll := NewCollector()

	staticSub, err := fs.Sub(StaticFS, "static")
	if err != nil {
		panic(fmt.Sprintf("debug static fs: %v", err))
	}
	srv := NewServer(ctrl, coll, staticSub)
	hook := newHook(ctrl, srv, coll)

	if hs, ok := inner.(hookSetter); ok {
		hs.SetFactoryHook(hook)
	}

	return &DebugFactory{
		inner:      inner,
		controller: ctrl,
		server:     srv,
		collector:  coll,
		hook:       hook,
	}
}

func (df *DebugFactory) StartServer() (string, error) {
	return df.server.Start()
}

func (df *DebugFactory) Close() {
	df.controller.Close()
}

// --- container.Factory delegation ---

func (df *DebugFactory) GetRegisteredComponents() map[string]any {
	return df.inner.GetRegisteredComponents()
}

func (df *DebugFactory) GetDefinitionRegistryPostProcessors() []container.DefinitionRegistryPostProcessor {
	return df.inner.GetDefinitionRegistryPostProcessors()
}

func (df *DebugFactory) SetRegistry(r container.SingletonRegistry) {
	df.inner.SetRegistry(r)
}

func (df *DebugFactory) SetConfigure(c configure.Configure) {
	df.inner.SetConfigure(c)
}

func (df *DebugFactory) PrepareComponents() error {
	return df.inner.PrepareComponents()
}

func (df *DebugFactory) Refresh() error {
	return df.inner.Refresh()
}

func (df *DebugFactory) GetComponents(opts ...container.Option) ([]any, error) {
	return df.inner.GetComponents(opts...)
}

func (df *DebugFactory) GetComponentByName(name string) (any, error) {
	return df.inner.GetComponentByName(name)
}

func (df *DebugFactory) GetConfigure() configure.Configure {
	return df.inner.GetConfigure()
}

func (df *DebugFactory) GetDefinitionRegistry() container.DefinitionRegistry {
	return df.inner.GetDefinitionRegistry()
}

func (df *DebugFactory) SetContext(ctx context.Context) {
	if cs, ok := df.inner.(contextSetter); ok {
		cs.SetContext(ctx)
	}
}
