package support

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/sync2"
)

type defaultDefinitionRegistry struct {
	metaMaps *sync2.Map[string, *component_definition.Meta]
}

func DefaultDefinitionRegistry() container.DefinitionRegistry {
	return &defaultDefinitionRegistry{
		metaMaps: sync2.New[string, *component_definition.Meta](),
	}
}

func (r *defaultDefinitionRegistry) RegisterMeta(m *component_definition.Meta) {
	r.metaMaps.Store(m.Name(), m)
	syslog.Pref("ComponentDefinitionRegistry").Tracef("register component definition for '%s'", m.Name())
}

func (r *defaultDefinitionRegistry) GetMetas(opts ...container.Option) []*component_definition.Meta {
	var metas = make([]*component_definition.Meta, 0)
	r.metaMaps.Range(func(k string, m *component_definition.Meta) bool {
		if container.And(opts...)(m) {
			metas = append(metas, m)
		}
		return true
	})
	return metas
}

func (r *defaultDefinitionRegistry) GetMetaByName(name string) *component_definition.Meta {
	if c, ok := r.metaMaps.Load(name); ok {
		return c
	}
	return nil
}

func (r *defaultDefinitionRegistry) GetMetaOrRegister(name string, component any) *component_definition.Meta {
	m, _ := r.metaMaps.LoadOrStoreFn(name, func() *component_definition.Meta {
		m := component_definition.NewMeta(component)
		m.SetName(name)
		return m
	})
	return m
}
