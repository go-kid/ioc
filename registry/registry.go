package registry

import (
	"fmt"
	"github.com/go-kid/ioc/scanner"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/util/list"
	"github.com/modern-go/concurrent"
)

/*
Registry
Dependency Register and Dependency Lookup
*/

type Registry interface {
	SetScanner(scanner scanner.Scanner)
	Register(cs ...any)
	GetComponents(opts ...Option) []*meta.Meta
	GetComponentByName(name string) *meta.Meta
	RemoveComponents(name string)
	ComponentInited(name string)
	IsComponentInited(name string) bool
}

type registry struct {
	scanner          scanner.Scanner
	components       *concurrent.Map
	initedComponents list.Set
}

func NewRegistry() Registry {
	return &registry{
		scanner:          scanner.New(),
		components:       concurrent.NewMap(),
		initedComponents: list.NewConcurrentSets(),
	}
}

func (r *registry) SetScanner(scanner scanner.Scanner) {
	r.scanner = scanner
	for _, m := range r.GetComponents() {
		r.components.Store(m.Name, r.scanner.ScanComponent(m.Raw))
	}
}

func (r *registry) Register(cs ...any) {
	for _, c := range cs {
		r.register(c)
	}
}

func (r *registry) register(c any) {
	if c == nil {
		panic("a nil value is passing to register")
	}
	var m *meta.Meta
	switch c.(type) {
	case *meta.Meta:
		m = c.(*meta.Meta)
	default:
		m = r.scanner.ScanComponent(c)
	}
	if a, ok := r.components.Load(m.Name); ok {
		ec := a.(*meta.Meta)
		if ec.Address == m.Address {
			return
		}
		panic(fmt.Sprintf("register duplicated component: %s", m.Name))
	}
	r.components.Store(m.Name, m)
}

func (r *registry) GetComponents(opts ...Option) []*meta.Meta {
	var metas = make([]*meta.Meta, 0)
	r.components.Range(func(k, v any) bool {
		m := v.(*meta.Meta)
		if accept(m, opts...) {
			metas = append(metas, m)
		}
		return true
	})
	return metas
}

func (r *registry) GetComponentByName(name string) *meta.Meta {
	if c, ok := r.components.Load(name); ok {
		return c.(*meta.Meta)
	}
	return nil
}

func (r *registry) RemoveComponents(name string) {
	r.components.Delete(name)
}

func (r *registry) IsComponentInited(name string) bool {
	return r.initedComponents.Exists(name)
}

func (r *registry) ComponentInited(name string) {
	r.initedComponents.Put(name)
}
