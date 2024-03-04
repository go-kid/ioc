package factory

import (
	"fmt"
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
)

type MetaFunc func(m *meta.Meta) error

type Factory interface {
	SetIfNilPostInitFunc(fn MetaFunc)
	Initialize(r registry.Registry, i injector.Injector, m *meta.Meta) error
}

type defaultFactory struct {
	postInitFunc MetaFunc
}

func Default() Factory {
	return &defaultFactory{}
}

func (f *defaultFactory) SetIfNilPostInitFunc(fn MetaFunc) {
	if f.postInitFunc == nil {
		f.postInitFunc = fn
	}
}

func (f *defaultFactory) Initialize(r registry.Registry, i injector.Injector, m *meta.Meta) error {
	syslog.Tracef("factory start initialize component %s", m.ID())
	if r.IsComponentInited(m.Name) {
		syslog.Tracef("component %s is already init, skip initialize", m.ID())
		return nil
	}

	syslog.Tracef("factory inject dependencies %s", m.ID())
	err := i.DependencyInject(r, m.ID(), m.AllDependencies())
	if err != nil {
		return fmt.Errorf("factory inject dependencies failed: %v", err)
	}

	r.ComponentInited(m.Name)

	for _, dependency := range m.AllDependencies() {
		for _, dm := range dependency.Injects {
			dm.DependBy(m)
			err := f.Initialize(r, i, dm)
			if err != nil {
				return err
			}
		}
	}

	if f.postInitFunc != nil {
		syslog.Tracef("factory do post init function")
		err = f.postInitFunc(m)
		if err != nil {
			return fmt.Errorf("factory do post init function failed: %v", err)
		}
	}

	syslog.Tracef("factory initialized component %s", m.ID())
	return nil
}
