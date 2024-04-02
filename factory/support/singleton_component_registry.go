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

func (r *defaultSingletonComponentRegistry) AddSingletonFactory(name string, method SingletonFactory) {
	r.singletonFactories.Store(name, method)
	r.logger().Tracef("add singleton factory for '%s' to singleton factories cache", name)
}

func (r *defaultSingletonComponentRegistry) RemoveSingleton(name string) {
	r.singletonObjects.Delete(name)
	r.earlySingletonObjects.Delete(name)
	r.singletonFactories.Delete(name)
	r.singletonCurrentlyInCreation.Remove(name)
	r.logger().Tracef("remove singleton '%s'", name)
}

func (r *defaultSingletonComponentRegistry) AddSingleton(name string, meta *component_definition.Meta) {
	r.singletonObjects.Store(name, meta)
	r.earlySingletonObjects.Delete(name)
	r.singletonFactories.Delete(name)
	r.logger().Tracef("put singleton '%s' to singleton objects", name)
	//syslog.Tracef("registry update component %s to inited", name)
}

func (r *defaultSingletonComponentRegistry) GetSingleton(name string, allowEarlyReference bool) (*component_definition.Meta, error) {
	// get component from inited components cache
	if meta, ok := r.singletonObjects.Load(name); ok {
		r.logger().Tracef("get singleton '%s' from singleton objects", name)
		return meta, nil
	}
	// get component from early export components cache
	if earlyComponent, ok := r.earlySingletonObjects.Load(name); ok {
		r.logger().Tracef("get singleton '%s' from early singleton cache", name)
		return earlyComponent, nil
	}
	if allowEarlyReference {
		// get component from singleton component factory cache
		if factory, ok := r.singletonFactories.Load(name); ok {
			r.logger().Tracef("get singleton '%s' from object factories", name)
			component, err := factory.GetComponent()
			if err != nil {
				return nil, err
			}
			r.earlySingletonObjects.Store(name, component)
			r.singletonFactories.Delete(name)
			r.logger().Tracef("singleton '%s' created from object factories and move to early cache", name)
			return component, nil
		}
	}
	return nil, nil
}

func (r *defaultSingletonComponentRegistry) GetSingletonOrCreateByFactory(name string, factory SingletonFactory) (*component_definition.Meta, error) {
	if singleton, loaded := r.singletonObjects.Load(name); loaded {
		return singleton, nil
	}
	r.logger().Tracef("singleton '%s' currently is new, start creating", name)
	r.singletonCurrentlyInCreation.Put(name)
	r.logger().Tracef("create instance of singleton '%s'", name)
	singleton, err := factory.GetComponent()
	if err != nil {
		return nil, err
	}
	r.logger().Tracef("singleton '%s' finished creating", name)
	r.singletonCurrentlyInCreation.Remove(name)
	r.AddSingleton(name, singleton)
	return singleton, nil
}

func (r *defaultSingletonComponentRegistry) IsSingletonCurrentlyInCreation(name string) bool {
	return r.singletonCurrentlyInCreation.Exists(name)
}

func (r *defaultSingletonComponentRegistry) logger() syslog.Logger {
	return syslog.GetLogger().Pref("SingletonComponentRegistry")
}
