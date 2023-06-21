package factory

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/meta"
	"github.com/go-kid/ioc/registry"
	"reflect"
)

type Factory interface {
	SetIfNilPreFunc(fn MetaFunc)
	SetIfNilPostInitFunc(fn MetaFunc)
	Initialize(r registry.Registry, m *meta.Meta) error
}

type MetaFunc func(m *meta.Meta) error

type DefaultFactory struct {
	preFunc      MetaFunc
	postInitFunc MetaFunc
}

func (f *DefaultFactory) SetIfNilPreFunc(fn MetaFunc) {
	if f.preFunc == nil {
		f.preFunc = fn
	}
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

	if f.preFunc != nil {
		err := f.preFunc(m)
		if err != nil {
			return err
		}
	}

	err := injector.DependencyInject(r, m)
	if err != nil {
		return err
	}
	r.ComponentInited(m.Name)

	for _, dependency := range m.Dependencies {
		switch dependency.Type.Kind() {
		case reflect.Slice:
			for i := 0; i < dependency.Value.Len(); i++ {
				elem := dependency.Value.Index(i)
				name := defination.GetComponentName(elem.Interface())
				dm := r.GetComponentByName(name)
				dm.DependBy(m)
				err := f.Initialize(r, dm)
				if err != nil {
					return err
				}
			}
		default:
			dm := r.GetComponentByName(dependency.Name())
			if dm == nil {
				return fmt.Errorf("component %s not found", dependency.Name())
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
