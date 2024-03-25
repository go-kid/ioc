package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/scanner"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/list"
	"github.com/go-kid/ioc/util/sync2"
)

var _ DefinitionRegistry = &container{}

type container struct {
	scanner                      scanner.Scanner
	metaMaps                     *sync2.Map[string, *component_definition.Meta]
	singletonObjects             *sync2.Map[string, *component_definition.Meta]
	earlySingletonObjects        *sync2.Map[string, *component_definition.Meta]
	singletonFactories           *sync2.Map[string, SingletonFactory]
	singletonCurrentlyInCreation list.Set
}

func DefaultDefinitionRegistry() DefinitionRegistry {
	return &container{
		metaMaps:                     sync2.New[string, *component_definition.Meta](),
		singletonObjects:             sync2.New[string, *component_definition.Meta](),
		earlySingletonObjects:        sync2.New[string, *component_definition.Meta](),
		singletonFactories:           sync2.New[string, SingletonFactory](),
		singletonCurrentlyInCreation: list.NewConcurrentSets(),
	}
}

func (r *container) GetComponentDefinitions(opts ...Option) []*component_definition.Meta {
	var metas = make([]*component_definition.Meta, 0)
	r.singletonObjects.Range(func(k string, m *component_definition.Meta) (shouldContinue bool) {
		if Accept(m, opts...) {
			metas = append(metas, m)
		}
		return true
	})
	return metas
}

func (r *container) GetComponentDefinitionByName(name string) (*component_definition.Meta, bool) {
	return r.singletonObjects.Load(name)
}

func (r *container) RegisterMeta(m *component_definition.Meta) {
	r.metaMaps.Store(m.Name, m)
}

func (r *container) GetMetas(opts ...Option) []*component_definition.Meta {
	var metas = make([]*component_definition.Meta, 0)
	r.metaMaps.Range(func(k string, m *component_definition.Meta) bool {
		if Accept(m, opts...) {
			metas = append(metas, m)
		}
		return true
	})
	return metas
}

func (r *container) GetMetaByName(name string) *component_definition.Meta {
	if c, ok := r.metaMaps.Load(name); ok {
		return c
	}
	return nil
}

func (r *container) AddSingletonFactory(name string, method SingletonFactory) {
	r.singletonFactories.Store(name, method)
	syslog.Tracef("definition registry add singleton factory for %s", name)
}

func (r *container) GetSingletonFactory(name string) (SingletonFactory, bool) {
	return r.singletonFactories.Load(name)
}

func (r *container) EarlyExportComponent(m *component_definition.Meta) {
	r.earlySingletonObjects.Store(m.Name, m)
	r.singletonFactories.Delete(m.Name)
}

func (r *container) GetEarlyExportComponent(name string) (*component_definition.Meta, bool) {
	return r.earlySingletonObjects.Load(name)
}

func (r *container) RemoveComponents(name string) {
	r.singletonObjects.Delete(name)
	syslog.Tracef("registry remove component %s", name)
}

func (r *container) ComponentInitialized(meta *component_definition.Meta) {
	r.singletonCurrentlyInCreation.Remove(meta.Name)
	r.singletonObjects.Store(meta.Name, meta)
	//syslog.Tracef("registry update component %s to inited", name)
}

func (r *container) GetComponents(opts ...Option) []any {
	var components = make([]any, 0)
	r.singletonObjects.Range(func(k string, m *component_definition.Meta) bool {
		if Accept(m, opts...) {
			components = append(components, m.Raw)
		}
		return true
	})
	return components
}

func (r *container) GetComponentByName(name string) any {
	if c, ok := r.singletonObjects.Load(name); ok {
		return c.Raw
	}
	return nil
}

func (r *container) GetComponent(name string) (*component_definition.Meta, error) {
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
	return nil, nil
}

func (r *container) BeforeSingletonCreation(name string) {
	r.singletonCurrentlyInCreation.Put(name)
}

func (r *container) IsSingletonCurrentlyInCreation(name string) bool {
	return r.singletonCurrentlyInCreation.Exists(name)
}
