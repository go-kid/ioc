package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/util/sync2"
)

var _ DefinitionRegistry = &defaultDefinitionRegistry{}

type defaultDefinitionRegistry struct {
	metaMaps *sync2.Map[string, *component_definition.Meta]
}

func DefaultDefinitionRegistry() DefinitionRegistry {
	return &defaultDefinitionRegistry{
		metaMaps: sync2.New[string, *component_definition.Meta](),
	}
}

func (r *defaultDefinitionRegistry) RegisterMeta(m *component_definition.Meta) {
	r.metaMaps.Store(m.Name, m)
}

func (r *defaultDefinitionRegistry) GetMetas(opts ...Option) []*component_definition.Meta {
	var metas = make([]*component_definition.Meta, 0)
	r.metaMaps.Range(func(k string, m *component_definition.Meta) bool {
		if Accept(m, opts...) {
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
