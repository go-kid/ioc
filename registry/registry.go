package registry

import (
	"fmt"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/sync2"
)

/*
Registry
Dependency Register and Dependency Lookup
*/

type SingletonRegistry interface {
	RegisterSingleton(name string, singleton any)
	//Register(cs ...any)
	GetSingleton(name string) (any, error)
	ContainsSingleton(name string) bool
	GetSingletonNames() []string
	GetSingletonCount() int
	//GetComponents(opts ...Option) []any
	//GetComponentByName(name string) any
	//RemoveComponents(name string)
}

type registry struct {
	componentsMap *sync2.Map[string, any]
	//metaMaps              *sync2.Map[string, *meta.Meta]
	//singletonObjects      *sync2.Map[string, *meta.Meta]
	//earlySingletonObjects *sync2.Map[string, *meta.Meta]
	//singletonFactories    *sync2.Map[string, FactoryMethod]
	//initedComponents      list.Set
}

func (r *registry) GetSingleton(name string) (any, error) {
	c, loaded := r.componentsMap.Load(name)
	if loaded {
		return c, nil
	}
	return nil, fmt.Errorf("singleton component %s not exist", name)
}

func (r *registry) ContainsSingleton(name string) bool {
	_, contains := r.componentsMap.Load(name)
	return contains
}

func (r *registry) GetSingletonNames() []string {
	var names []string
	r.componentsMap.Range(func(key string, _ any) (shouldContinue bool) {
		names = append(names, key)
		return true
	})
	return names
}

func (r *registry) GetSingletonCount() int {
	return len(r.GetSingletonNames())
}

func NewRegistry() SingletonRegistry {
	return &registry{
		componentsMap: sync2.New[string, any](),
		//metaMaps:              sync2.New[string, *meta.Meta](),
		//singletonObjects:      sync2.New[string, *meta.Meta](),
		//earlySingletonObjects: sync2.New[string, *meta.Meta](),
		//singletonFactories:    sync2.New[string, FactoryMethod](),
		//initedComponents:      list.NewConcurrentSets(),
	}
}

func (r *registry) RegisterSingleton(name string, singleton any) {
	if exist, loaded := r.componentsMap.Load(name); loaded {
		if exist != singleton {
			syslog.Panicf("register duplicated component %s", name)
		}
		return
	}
	r.componentsMap.Store(name, singleton)
	syslog.Tracef("singleton registry register component %s", name)
}

//func (r *registry) Register(cs ...any) {
//	if len(cs) < 1 {
//		return
//	}
//	for _, c := range cs {
//		if fas.IsNil(c) {
//			syslog.Panicf("register a nil value component %s", reflectx.Id(c))
//		}
//		name, _ := component_defination.GetComponentName(c)
//		if lc, loaded := r.componentsMap.Load(name); loaded {
//			if lc != c {
//				syslog.Panicf("register duplicated component %s", name)
//			}
//		}
//		r.componentsMap.Store(name, c)
//	}
//	//for _, c := range cs {
//	//r.scanAndCache(r.sc, c)
//	//}
//	r.components = append(r.components, cs...)
//}
//
//func (r *registry) ScanAllComponents() []*meta.Meta {
//	var metas = make([]*meta.Meta, len(r.components))
//	for i, component := range r.components {
//		metas[i] = r.sc.ScanComponent(component)
//	}
//	return metas
//}

//func (r *registry) Scan(sc scanner.Scanner) {
//	syslog.Tracef("registry scan by scanner %s", reflectx.Id(sc))
//	wg := sync.WaitGroup{}
//	wg.Add(len(r.components))
//	for _, component := range r.components {
//		go func(c any) {
//			r.scanAndCache(sc, c)
//			wg.Done()
//		}(component)
//	}
//	wg.Wait()
//	syslog.Trace("registry scan finished")
//}

//func (r *registry) scanAndCache(c any) {
//	var m *meta.Meta
//	switch c.(type) {
//	case *meta.Meta:
//		m = c.(*meta.Meta)
//		syslog.Tracef("registry scan raw meta %s", m.ID())
//	default:
//		m = r.sc.ScanComponent(c)
//		syslog.Tracef("registry scan component %s", m.ID())
//	}
//	if a, ok := r.componentsMap.Load(m.Name); ok {
//		if a.ID() != m.ID() {
//			syslog.Panicf("register duplicated component, cached: %s, new register: %s", a.ID(), m.ID())
//		}
//		return
//	}
//	r.metaMaps.Store(m.Name, m)
//	syslog.Tracef("registry cache component %s", m.ID())
//}

//func (r *registry) GetMetas(opts ...Option) []*meta.Meta {
//	var metas = make([]*meta.Meta, 0)
//	r.metaMaps.Range(func(k string, m *meta.Meta) bool {
//		if Accept(m, opts...) {
//			metas = append(metas, m)
//		}
//		return true
//	})
//	return metas
//}
//
//func (r *registry) GetMetaByName(name string) *meta.Meta {
//	if c, ok := r.metaMaps.Load(name); ok {
//		return c
//	}
//	return nil
//}
//
//func (r *registry) GetComponents(opts ...Option) []any {
//	var components = make([]any, 0)
//	r.singletonObjects.Range(func(k string, m *meta.Meta) bool {
//		if Accept(m, opts...) {
//			components = append(components, m.Raw)
//		}
//		return true
//	})
//	return components
//}
//
//func (r *registry) GetComponentByName(name string) any {
//	if c, ok := r.singletonObjects.Load(name); ok {
//		return c.Raw
//	}
//	return nil
//}
//
//func (r *registry) SetSingletonFactoryMethod(name string, method FactoryMethod) {
//	r.singletonFactories.Store(name, method)
//}
//
//func (r *registry) GetSingletonFactoryMethod(name string) (FactoryMethod, bool) {
//	return r.singletonFactories.Load(name)
//}
//
//func (r *registry) EarlyExportComponent(m *meta.Meta) {
//	r.earlySingletonObjects.Store(m.Name, m)
//	r.singletonFactories.Delete(m.Name)
//}
//
//func (r *registry) GetEarlyExportComponent(name string) (*meta.Meta, bool) {
//	return r.earlySingletonObjects.Load(name)
//}
//
//func (r *registry) RemoveComponents(name string) {
//	r.singletonObjects.Delete(name)
//	syslog.Tracef("registry remove component %s", name)
//}
//
//func (r *registry) IsComponentInited(name string) bool {
//	_, loaded := r.singletonObjects.Load(name)
//	return loaded
//}
//
//func (r *registry) ComponentInited(name string) error {
//	m, loaded := r.earlySingletonObjects.Load(name)
//	if !loaded {
//		return fmt.Errorf("component %s is not initiated", name)
//	}
//	r.singletonObjects.Store(name, m)
//	r.earlySingletonObjects.Delete(name)
//	//syslog.Tracef("registry update component %s to inited", name)
//	return nil
//}
