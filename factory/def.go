package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/registry"
)

type MetaFunc func(m *component_definition.Meta) error

type Factory interface {
	SetRegistry(r registry.Registry)
	SetConfigure(c configure.Configure)
	AddInjectionRules(rules ...InjectionRule)
	PrepareSpecialComponents() error
	Initialize(metas ...*component_definition.Meta) error
}

type InjectionRule interface {
	RuleName() string
	Priority() int
	Condition(d *component_definition.Node) bool
	Candidates(r registry.Registry, d *component_definition.Node) ([]*component_definition.Meta, error)
}
