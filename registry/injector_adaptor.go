package registry

import (
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/meta"
	"github.com/samber/lo"
	"reflect"
)

type registryInjector struct {
	Registry
}

func newRegistryInjector(r Registry) injector.Injector {
	return &registryInjector{r}
}

func (r *registryInjector) GetOneByInterfaceType(typ reflect.Type) (reflect.Value, bool) {
	metas := r.GetBeansByInterfaceType(typ)
	if len(metas) < 1 {
		return reflect.Value{}, false
	}
	var v = metas[0].Value
	for _, m := range metas {
		if _, ok := m.Raw.(defination.NamingComponent); !ok {
			v = m.Value
			break
		}
	}
	return v, true
}

func (r *registryInjector) GetsByInterfaceType(typ reflect.Type) []reflect.Value {
	metas := r.GetBeansByInterfaceType(typ)
	return lo.Map(metas, func(item *meta.Meta, _ int) reflect.Value {
		return item.Value
	})
}

func (r *registryInjector) GetByName(name string) (reflect.Value, bool) {
	m := r.GetComponentByName(name)
	if m == nil {
		return reflect.Value{}, false
	}
	return m.Value, true
}
