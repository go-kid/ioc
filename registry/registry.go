package registry

import (
	"github.com/kid-hash/kid-ioc/meta"
	"github.com/kid-hash/kid-ioc/util/list"
	"github.com/kid-hash/kid-ioc/util/reflectx"
	"github.com/modern-go/concurrent"
	"reflect"
)

/*
Registry
Dependency Register and Dependency Lookup
*/
type Registry struct {
	components       *concurrent.Map
	initedComponents *list.ConcurrentSets
}

func NewRegistry() *Registry {
	return &Registry{
		components:       concurrent.NewMap(),
		initedComponents: list.NewConcurrentSets(),
	}
}

func (r *Registry) Register(cs ...interface{}) {
	for _, c := range cs {
		r.register(c)
	}
}

func (r *Registry) register(c interface{}) {
	if c == nil {
		panic("a nil value is passing to register")
	}
	m := meta.NewMeta(c)
	if a, ok := r.components.Load(m.Name); ok {
		ec := a.(*meta.Meta)
		if ec.Address == m.Address {
			return
		}
		//log.Panic().Msgf("register duplicated component: %s", m.Name)
	}
	r.components.Store(m.Name, m)
	//fmt.Println("store:", m.Name)
}

func (r *Registry) GetComponents() []*meta.Meta {
	var metas = make([]*meta.Meta, 0)
	r.components.Range(func(k, v interface{}) bool {
		metas = append(metas, v.(*meta.Meta))
		return true
	})
	return metas
}

func (r *Registry) GetComponentByName(name string) *meta.Meta {
	if c, ok := r.components.Load(name); ok {
		return c.(*meta.Meta)
	}
	return nil
}

func (r *Registry) GetBeansByInterfaceType(typ reflect.Type) []*meta.Meta {
	return r.GetBeansByInterface(reflect.New(typ).Interface())
}

func (r *Registry) GetBeansByInterface(a interface{}) []*meta.Meta {
	var beans = make([]*meta.Meta, 0)
	for _, m := range r.GetComponents() {
		if reflectx.IsTypeImplement(m.Type, a) {
			beans = append(beans, m)
		}
	}
	return beans
}

func (r *Registry) RemoveComponents(name string) {
	r.components.Delete(name)
}

func (r *Registry) IsComponentInited(name string) bool {
	return r.initedComponents.Exists(name)
}

func (r *Registry) ComponentInited(name string) {
	r.initedComponents.Put(name)
}
