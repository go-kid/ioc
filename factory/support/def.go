package support

import "github.com/go-kid/ioc/component_definition"

type SingletonRegistry interface {
	RegisterSingleton(singleton any)
	GetSingleton(name string) (any, error)
	ContainsSingleton(name string) bool
	GetSingletonNames() []string
	GetSingletonCount() int
}

type RegisterMeta func() *component_definition.Meta

type DefinitionRegistry interface {
	RegisterMeta(m *component_definition.Meta)
	GetMetas(opts ...Option) []*component_definition.Meta
	GetMetaByName(name string) *component_definition.Meta
	GetMetaOrRegister(name string, handler RegisterMeta) *component_definition.Meta
}

type SingletonComponentRegistry interface {
	AddSingleton(name string, meta *component_definition.Meta)
	AddSingletonFactory(name string, method SingletonFactory)
	GetSingleton(name string, allowEarlyReference bool) (*component_definition.Meta, error)
	RemoveSingleton(name string)
	GetSingletonOrCreateByFactory(name string, factory SingletonFactory) (*component_definition.Meta, error)
	IsSingletonCurrentlyInCreation(name string) bool
}

type SingletonFactory interface {
	GetComponent() (*component_definition.Meta, error)
}

type FuncSingletonFactory func() (*component_definition.Meta, error)

func (d FuncSingletonFactory) GetComponent() (*component_definition.Meta, error) {
	return d()
}
