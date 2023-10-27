package registry

import (
	"fmt"
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/scanner"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/util/list"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/modern-go/concurrent"
	"github.com/samber/lo"
	"reflect"
)

/*
Registry
Dependency Register and Dependency Lookup
*/

type Registry interface {
	SetScanner(scanner scanner.Scanner)
	Register(cs ...any)
	GetComponents() []*meta.Meta
	GetComponentByName(name string) *meta.Meta
	GetBeansByInterfaceType(typ reflect.Type) []*meta.Meta
	GetBeansByInterface(a any) []*meta.Meta
	RemoveComponents(name string)
	ComponentInited(name string)
	IsComponentInited(name string) bool
	GetComponentsByFunc(funcName string) []*meta.Meta
	Injector() injector.Injector
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

func (r *registry) GetComponents() []*meta.Meta {
	var metas = make([]*meta.Meta, 0)
	r.components.Range(func(k, v any) bool {
		metas = append(metas, v.(*meta.Meta))
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

func (r *registry) GetBeansByInterfaceType(typ reflect.Type) []*meta.Meta {
	return r.GetBeansByInterface(reflect.New(typ).Interface())
}

func (r *registry) GetBeansByInterface(a any) []*meta.Meta {
	var beans = make([]*meta.Meta, 0)
	for _, m := range r.GetComponents() {
		if reflectx.IsTypeImplement(m.Type, a) {
			beans = append(beans, m)
		}
	}
	return beans
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

func (r *registry) GetComponentsByFunc(funcName string) []*meta.Meta {
	return lo.Filter(r.GetComponents(), func(item *meta.Meta, _ int) bool {
		return item.Value.MethodByName(funcName).IsValid()
	})
}

func (r *registry) Injector() injector.Injector {
	return newRegistryInjector(r)
}
