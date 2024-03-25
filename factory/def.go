package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner"
)

type MetaFunc func(m *component_definition.Meta) error

type Factory interface {
	SetRegistry(r registry.SingletonRegistry)
	SetConfigure(c configure.Configure)
	SetScanner(sc scanner.Scanner)
	AddInjectionRules(rules ...InjectionRule)
	PrepareComponents() error
	Initialize() error
	GetComponents(opts ...Option) []any
	GetComponentByName(name string) (any, error)
}

type InjectionRule interface {
	RuleName() string
	Priority() int
	Condition(d *component_definition.Node) bool
	Candidates(r DefinitionRegistry, d *component_definition.Node) ([]*component_definition.Meta, error)
}

type DefinitionRegistry interface {
	RegisterMeta(m *component_definition.Meta)
	ComponentInitialized(meta *component_definition.Meta)
	AddSingletonFactory(name string, method SingletonFactory)
	GetSingletonFactory(name string) (SingletonFactory, bool)
	EarlyExportComponent(m *component_definition.Meta)
	GetEarlyExportComponent(name string) (*component_definition.Meta, bool)
	GetMetas(opts ...Option) []*component_definition.Meta
	GetMetaByName(name string) *component_definition.Meta
	GetComponentDefinitions(opts ...Option) []*component_definition.Meta
	GetComponentDefinitionByName(name string) (*component_definition.Meta, bool)
	GetComponent(name string) (*component_definition.Meta, error)
	BeforeSingletonCreation(name string)
	IsSingletonCurrentlyInCreation(name string) bool
}

type SingletonFactory interface {
	GetComponent() (*component_definition.Meta, error)
}

type FuncSingletonFactory func() (*component_definition.Meta, error)

func (d FuncSingletonFactory) GetComponent() (*component_definition.Meta, error) {
	return d()
}
