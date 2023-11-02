package registry

import (
	"github.com/go-kid/ioc/scanner"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/util/list"
	"github.com/modern-go/concurrent"
	"log"
	"sync"
)

/*
Registry
Dependency Register and Dependency Lookup
*/

type Registry interface {
	Register(cs ...any)
	SetScanner(scanner scanner.Scanner)
	SetExtraTags(tags ...string)
	Scan()
	GetComponents(opts ...Option) []*meta.Meta
	GetComponentByName(name string) *meta.Meta
	RemoveComponents(name string)
	ComponentInited(name string)
	IsComponentInited(name string) bool
}

type registry struct {
	scanner          scanner.Scanner
	components       []any
	metaMaps         *concurrent.Map
	initedComponents list.Set
}

func NewRegistry() Registry {
	return &registry{
		scanner:          scanner.New(),
		metaMaps:         concurrent.NewMap(),
		initedComponents: list.NewConcurrentSets(),
	}
}

func (r *registry) Register(cs ...any) {
	if len(cs) < 1 {
		return
	}
	for _, c := range cs {
		if c == nil {
			panic("a nil value is passing to register")
		}
	}
	r.components = append(r.components, cs...)
}

func (r *registry) SetScanner(scanner scanner.Scanner) {
	r.scanner = scanner
}

func (r *registry) SetExtraTags(tags ...string) {
	r.scanner.AddTags(tags...)
}

func (r *registry) Scan() {
	wg := sync.WaitGroup{}
	wg.Add(len(r.components))
	for _, component := range r.components {
		go func(c any) {
			r.register(c)
			wg.Done()
		}(component)
	}
	wg.Wait()
}

func (r *registry) register(c any) {
	var m *meta.Meta
	switch c.(type) {
	case *meta.Meta:
		m = c.(*meta.Meta)
	default:
		m = r.scanner.ScanComponent(c)
	}
	if a, ok := r.metaMaps.Load(m.Name); ok {
		ec := a.(*meta.Meta)
		if ec.Address == m.Address {
			return
		}
		log.Fatalf("register duplicated component: %s", m.Name)
	}
	r.metaMaps.Store(m.Name, m)
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
}

func (r *registry) IsComponentInited(name string) bool {
	return r.initedComponents.Exists(name)
}

func (r *registry) ComponentInited(name string) {
	r.initedComponents.Put(name)
}
