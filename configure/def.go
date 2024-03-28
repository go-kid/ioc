package configure

import (
	"github.com/go-kid/ioc/component_definition"
)

type Loader interface {
	LoadConfig() ([]byte, error)
}

type Binder interface {
	SetConfig(c []byte) error
	Get(path string) any
	Set(path string, val any)
	Unmarshall(key string, a any) error
}

type Configure interface {
	Binder
	AddLoaders(loaders ...Loader)
	SetLoaders(loaders ...Loader)
	AddPopulateProcessors(processors ...PopulateProcessor)
	SetBinder(binder Binder)
	Initialize() error
	PopulateProperties(metas ...*component_definition.Meta) error
}

type PopulateProcessor interface {
	Order() int
	Filter(d *component_definition.Node) bool
	Populate(r Binder, d *component_definition.Node) error
}
