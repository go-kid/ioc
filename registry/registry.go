package registry

import (
	"fmt"
	"github.com/go-kid/ioc/meta"
	"github.com/go-kid/ioc/util/list"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/modern-go/concurrent"
	"reflect"
)

/*
Registry
Dependency Register and Dependency Lookup
*/

type Registry interface {
	Register(cs ...interface{})
	GetComponents() []*meta.Meta
	GetComponentByName(name string) *meta.Meta
	GetBeansByInterfaceType(typ reflect.Type) []*meta.Meta
	GetBeansByInterface(a interface{}) []*meta.Meta
	RemoveComponents(name string)
	ComponentInited(name string)
	IsComponentInited(name string) bool
}

type registry struct {
	components       *concurrent.Map
	initedComponents list.Set
}

func NewRegistry() Registry {
	return &registry{
		components:       concurrent.NewMap(),
		initedComponents: list.NewConcurrentSets(),
	}
}

func (r *registry) Register(cs ...interface{}) {
	for _, c := range cs {
		r.register(c)
	}
}

func (r *registry) register(c interface{}) {
	if c == nil {
		panic("a nil value is passing to register")
	}
	var m *meta.Meta
	switch c.(type) {
	case *meta.Meta:
		m = c.(*meta.Meta)
	default:
		m = meta.NewMeta(c)
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
	r.components.Range(func(k, v interface{}) bool {
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

func (r *registry) GetBeansByInterface(a interface{}) []*meta.Meta {
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
