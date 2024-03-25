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
	PrepareComponents() error
	Initialize() error
}

type InjectionRule interface {
	RuleName() string
	Priority() int
	Condition(d *meta.Node) bool
	Candidates(r BuildContainer, d *meta.Node) ([]*meta.Meta, error)
}

type BuildContainer interface {
	IsComponentInited(name string) bool
	ComponentInited(name string) error
	SetSingletonFactoryMethod(name string, method FactoryMethod)
	GetSingletonFactoryMethod(name string) (FactoryMethod, bool)
	EarlyExportComponent(m *meta.Meta)
	GetEarlyExportComponent(name string) (*meta.Meta, bool)
	GetMetas(opts ...Option) []*meta.Meta
	GetMetaByName(name string) *meta.Meta
}
