package factory

import (
	"fmt"
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/meta"
	"github.com/go-kid/ioc/registry"
	"reflect"
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

	err := injector.DependencyInject(r.Injector(), m.ID(), m.Dependencies)
	if err != nil {
		return err
	}
	r.ComponentInited(m.Name)

	for _, dependency := range m.Dependencies {
		switch dependency.Type.Kind() {
		case reflect.Slice:
			for i := 0; i < dependency.Value.Len(); i++ {
				elem := dependency.Value.Index(i)
				name := injector.GetComponentName(elem)
				dm := r.GetComponentByName(name)
				if dm == nil {
					return fmt.Errorf("component %s not found", dependency.Id())
				}
				dm.DependBy(m)
				err := f.Initialize(r, dm)
				if err != nil {
					return err
				}
			}
		default:
			dm := r.GetComponentByName(dependency.Id())
			if dm == nil {
				return fmt.Errorf("component %s not found", dependency.Id())
			}
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
