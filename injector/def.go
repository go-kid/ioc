package injector

import (
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
)

type InjectProcessor interface {
	RuleName() string
	Priority() int
	Filter(d *meta.Node) bool
	Inject(r registry.Registry, d *meta.Node) error
}

type Injector interface {
	AddCustomizedInjectors(ips ...InjectProcessor)
	DependencyInject(r registry.Registry, id string, dependencies []*meta.Node) error
}
