package factory

import (
	"fmt"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/list"
	"github.com/go-kid/ioc/util/sync2"
)

var _ SingletonComponentRegistry = &defaultSingletonComponentRegistry{}

type defaultSingletonComponentRegistry struct {
	singletonObjects                 *sync2.Map[string, any]
	earlySingletonObjects            *sync2.Map[string, any]
	singletonFactories               *sync2.Map[string, SingletonFactory]
	singletonCurrentlyInCreation     list.Set
	singletonsCurrentlyInDestruction list.Set
}

func DefaultSingletonComponentRegistry() SingletonComponentRegistry {
	return &defaultSingletonComponentRegistry{
		singletonObjects:                 sync2.New[string, any](),
		earlySingletonObjects:            sync2.New[string, any](),
		singletonFactories:               sync2.New[string, SingletonFactory](),
		singletonCurrentlyInCreation:     list.NewConcurrentSets(),
		singletonsCurrentlyInDestruction: list.NewConcurrentSets(),
	}
}

func (r *defaultSingletonComponentRegistry) GetSingletonByFactory(name string, factory SingletonFactory) (any, error) {
	singleton, ok := r.singletonObjects.Load(name)
	if !ok {
		if r.singletonsCurrentlyInDestruction.Exists(name) {
			return nil, fmt.Errorf("singleton creation not allowed while singleton is in destruction")
		}
		err := r.BeforeSingletonCreation(name)
		if err != nil {
			return nil, err
		}
		singleton, err = factory.GetComponent()
		if err != nil {
			if singleton, ok = r.singletonObjects.Load(name); ok {
				return singleton, nil
			}
			return nil, err
		}
		err = r.AfterSingletonCreation(name)
		if err != nil {
			return nil, err
		}
		r.AddSingleton(name, singleton)
	}
	return singleton, nil
}

func (r *defaultSingletonComponentRegistry) AddSingletonFactory(name string, method SingletonFactory) {
	r.singletonFactories.Store(name, method)
	syslog.Tracef("definition registry add singleton factory for %s", name)
}

func (r *defaultSingletonComponentRegistry) RemoveComponents(name string) {
	r.singletonObjects.Delete(name)
	syslog.Tracef("registry remove component %s", name)
}

func (r *defaultSingletonComponentRegistry) AddSingleton(name string, meta any) {
	r.singletonCurrentlyInCreation.Remove(name)
	r.singletonObjects.Store(name, meta)
	//syslog.Tracef("registry update component %s to inited", name)
}

func (r *defaultSingletonComponentRegistry) GetSingleton(name string) (any, error) {
	// get component from inited components cache
	var singleton any
	if meta, ok := r.singletonObjects.Load(name); ok {
		syslog.Tracef("definition registry get component definition by name %s", name)
		singleton = meta
	} else if r.singletonCurrentlyInCreation.Exists(name) {
		// get component from early export components cache
		if earlyComponent, ok := r.earlySingletonObjects.Load(name); ok {
			syslog.Tracef("definition registry get early export component %s", name)
			singleton = earlyComponent
		} else if factory, ok := r.singletonFactories.Load(name); ok {
			// get component from singleton component factory cache
			syslog.Tracef("definition registry get singleton factory %s", name)
			component, err := factory.GetComponent()
			if err != nil {
				return nil, err
			}
			r.earlySingletonObjects.Store(name, component)
			r.singletonFactories.Delete(name)
			return component, nil
		}
	}

	return singleton, nil
}

func (r *defaultSingletonComponentRegistry) BeforeSingletonCreation(name string) error {
	if r.singletonCurrentlyInCreation.Exists(name) {
		return fmt.Errorf("singleton '%s' is currently in creation", name)
	}
	r.singletonCurrentlyInCreation.Put(name)
	return nil
}

func (r *defaultSingletonComponentRegistry) IsSingletonCurrentlyInCreation(name string) bool {
	return r.singletonCurrentlyInCreation.Exists(name)
}

func (r *defaultSingletonComponentRegistry) AfterSingletonCreation(name string) error {
	if !r.singletonCurrentlyInCreation.Exists(name) {
		return fmt.Errorf("singleton '%s' isn't in creation", name)
	}
	r.singletonCurrentlyInCreation.Remove(name)
	return nil
}
