package support

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/list"
	"github.com/go-kid/ioc/util/sync2"
)

type defaultSingletonComponentRegistry struct {
	singletonObjects             *sync2.Map[string, *component_definition.Meta]
	earlySingletonObjects        *sync2.Map[string, *component_definition.Meta]
	singletonFactories           *sync2.Map[string, SingletonFactory]
	singletonCurrentlyInCreation list.Set
}

func DefaultSingletonComponentRegistry() SingletonComponentRegistry {
	return &defaultSingletonComponentRegistry{
		singletonObjects:             sync2.New[string, *component_definition.Meta](),
		earlySingletonObjects:        sync2.New[string, *component_definition.Meta](),
		singletonFactories:           sync2.New[string, SingletonFactory](),
		singletonCurrentlyInCreation: list.NewConcurrentSets(),
	}
}

func (r *defaultSingletonComponentRegistry) GetComponentDefinitions(opts ...Option) []*component_definition.Meta {
	var metas = make([]*component_definition.Meta, 0)
	r.singletonObjects.Range(func(k string, m *component_definition.Meta) (shouldContinue bool) {
		if Accept(m, opts...) {
			metas = append(metas, m)
		}
		return true
	})
	return metas
}

func (r *defaultSingletonComponentRegistry) GetComponentDefinitionByName(name string) (*component_definition.Meta, bool) {
	return r.singletonObjects.Load(name)
}

func (r *defaultSingletonComponentRegistry) AddSingletonFactory(name string, method SingletonFactory) {
	r.singletonFactories.Store(name, method)
	syslog.Tracef("definition registry add singleton factory for %s", name)
}

func (r *defaultSingletonComponentRegistry) RemoveComponents(name string) {
	r.singletonObjects.Delete(name)
	syslog.Tracef("registry remove component %s", name)
}

func (r *defaultSingletonComponentRegistry) AddSingleton(name string, meta *component_definition.Meta) {
	r.singletonObjects.Store(name, meta)
	r.earlySingletonObjects.Delete(name)
	r.singletonFactories.Delete(name)
	//syslog.Tracef("registry update component %s to inited", name)
}

func (r *defaultSingletonComponentRegistry) GetSingleton(name string, allowEarlyReference bool) (*component_definition.Meta, error) {
	// get component from inited components cache
	if meta, ok := r.singletonObjects.Load(name); ok {
		syslog.Tracef("definition registry get component definition by name %s", name)
		return meta, nil
	}
	// get component from early export components cache
	if earlyComponent, ok := r.earlySingletonObjects.Load(name); ok {
		syslog.Tracef("definition registry get early export component %s", name)
		return earlyComponent, nil
	}
	if allowEarlyReference {
		// get component from singleton component factory cache
		if factory, ok := r.singletonFactories.Load(name); ok {
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
	return nil, nil
}

func (r *defaultSingletonComponentRegistry) GetSingletonOrCreateByFactory(name string, factory SingletonFactory) (*component_definition.Meta, error) {
	if singleton, loaded := r.singletonObjects.Load(name); loaded {
		return singleton, nil
	}
	r.singletonCurrentlyInCreation.Put(name)
	singleton, err := factory.GetComponent()
	if err != nil {
		return nil, err
	}
	r.singletonCurrentlyInCreation.Remove(name)
	r.AddSingleton(name, singleton)
	return singleton, nil
}

func (r *defaultSingletonComponentRegistry) BeforeSingletonCreation(name string) {
	r.singletonCurrentlyInCreation.Put(name)
}

func (r *defaultSingletonComponentRegistry) IsSingletonCurrentlyInCreation(name string) bool {
	return r.singletonCurrentlyInCreation.Exists(name)
}
