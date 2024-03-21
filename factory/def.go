package factory

import (
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
)

type MetaFunc func(m *meta.Meta) error

type Factory interface {
	SetRegistry(r registry.Registry)
	SetConfigure(c configure.Configure)
	AddInjectionRules(rules ...InjectionRule)
	PrepareSpecialComponents() error
	Initialize(metas ...*meta.Meta) error
}

type InjectionRule interface {
	RuleName() string
	Priority() int
	Condition(d *meta.Node) bool
	Candidates(r registry.Registry, d *meta.Node) ([]*meta.Meta, error)
}
