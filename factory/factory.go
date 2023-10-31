package factory

import (
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
)

type Factory interface {
	SetIfNilPostInitFunc(fn MetaFunc)
	Initialize(r registry.Registry, m *meta.Meta) error
}

type MetaFunc func(m *meta.Meta) error

type DefaultFactory struct {
	postInitFunc MetaFunc
}

func (f *DefaultFactory) SetIfNilPostInitFunc(fn MetaFunc) {
	if f.postInitFunc == nil {
		f.postInitFunc = fn
	}
}

func (f *DefaultFactory) Initialize(r registry.Registry, m *meta.Meta) error {
	if r.IsComponentInited(m.Name) {
		return nil
	}

	err := injector.DependencyInject(r, m.ID(), m.AllDependencies())
	if err != nil {
		return err
	}

	r.ComponentInited(m.Name)

	for _, dependency := range m.AllDependencies() {
		for _, dm := range dependency.Injects {
			dm.DependBy(m)
			err := f.Initialize(r, dm)
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
