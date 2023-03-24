package ioc

import (
	"github.com/kid-hash/kid-ioc/util/list"
	"github.com/kid-hash/kid-ioc/util/reflectx"
	"github.com/modern-go/concurrent"
	"reflect"
)

/*
registry
Dependency Register and Dependency Lookup
*/
type registry struct {
	components       *concurrent.Map
	initedComponents *list.ConcurrentSets
}

func NewRegistry() *registry {
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
	m := newMeta(c)
	if a, ok := r.components.Load(m.Name); ok {
		ec := a.(*meta)
		if ec.Address == m.Address {
			return
		}
		//log.Panic().Msgf("register duplicated component: %s", m.Name)
	}
	r.components.Store(m.Name, m)
	//fmt.Println("store:", m.Name)
}

func (r *registry) GetComponents() []*meta {
	var metas = make([]*meta, 0)
	r.components.Range(func(k, v interface{}) bool {
		metas = append(metas, v.(*meta))
		return true
	})
	return metas
}

func (r *registry) GetComponentByName(name string) *meta {
	if c, ok := r.components.Load(name); ok {
		return c.(*meta)
	}
	return nil
}

func (r *registry) GetBeansByInterfaceType(typ reflect.Type) []*meta {
	return r.GetBeansByInterface(reflect.New(typ).Interface())
}

func (r *registry) GetBeansByInterface(a interface{}) []*meta {
	var beans = make([]*meta, 0)
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
