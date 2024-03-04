package registry

import (
	"github.com/go-kid/ioc/scanner"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/fas"
	"github.com/go-kid/ioc/util/list"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/modern-go/concurrent"
	"sync"
)

/*
Registry
Dependency Register and Dependency Lookup
*/

type Registry interface {
	Scan(sc scanner.Scanner)
	Register(cs ...any)
	GetComponents(opts ...Option) []*meta.Meta
	GetComponentByName(name string) *meta.Meta
	RemoveComponents(name string)
	ComponentInited(name string)
	IsComponentInited(name string) bool
}

type registry struct {
	components       []any
	metaMaps         *concurrent.Map
	initedComponents list.Set
}

func NewRegistry() Registry {
	return &registry{
		metaMaps:         concurrent.NewMap(),
		initedComponents: list.NewConcurrentSets(),
	}
}

func (r *registry) Register(cs ...any) {
	if len(cs) < 1 {
		return
	}
	for _, c := range cs {
		if fas.IsNil(c) {
			syslog.Panicf("register a nil value component %s", reflectx.Id(c))
		}
	}
	r.components = append(r.components, cs...)
}

func (r *registry) Scan(sc scanner.Scanner) {
	syslog.Tracef("registry scan by scanner %s", reflectx.Id(sc))
	wg := sync.WaitGroup{}
	wg.Add(len(r.components))
	for _, component := range r.components {
		go func(c any) {
			r.scanAndCache(sc, c)
			wg.Done()
		}(component)
	}
	wg.Wait()
	syslog.Trace("registry scan finished")
}

func (r *registry) scanAndCache(sc scanner.Scanner, c any) {
	var m *meta.Meta
	switch c.(type) {
	case *meta.Meta:
		m = c.(*meta.Meta)
		syslog.Tracef("registry scan raw meta %s", m.ID())
	default:
		m = sc.ScanComponent(c)
		syslog.Tracef("registry scan component %s", m.ID())
	}
	if a, ok := r.metaMaps.Load(m.Name); ok {
		if a.(*meta.Meta).ID() != m.ID() {
			syslog.Panicf("register duplicated component %s", m.Name)
		}
		return
	}
	r.metaMaps.Store(m.Name, m)
	syslog.Tracef("registry cache component %s", m.ID())
}

func (r *registry) GetComponents(opts ...Option) []*meta.Meta {
	var metas = make([]*meta.Meta, 0)
	r.metaMaps.Range(func(k, v any) bool {
		m := v.(*meta.Meta)
		if accept(m, opts...) {
			metas = append(metas, m)
		}
		return true
	})
	return metas
}

func (r *registry) GetComponentByName(name string) *meta.Meta {
	if c, ok := r.metaMaps.Load(name); ok {
		return c.(*meta.Meta)
	}
	return nil
}

func (r *registry) RemoveComponents(name string) {
	r.metaMaps.Delete(name)
	syslog.Tracef("registry remove component %s", name)
}

func (r *registry) IsComponentInited(name string) bool {
	return r.initedComponents.Exists(name)
}

func (r *registry) ComponentInited(name string) {
	r.initedComponents.Put(name)
	syslog.Tracef("registry update component %s to inited", name)
}
