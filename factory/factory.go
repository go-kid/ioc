package factory

import (
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
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
	if r.IsComponentInited(m.Name) {
		return nil
	}

	err := i.DependencyInject(r, m.ID(), m.AllDependencies())
	if err != nil {
		return err
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
		err = f.postInitFunc(m)
		if err != nil {
			return err
		}
	}

	return nil
}
