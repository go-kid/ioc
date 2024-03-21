package configure

import (
	"github.com/go-kid/ioc/scanner/meta"
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
	PopulateProperties(metas ...*meta.Meta) error
}

type PopulateProcessor interface {
	Order() int
	Filter(d *meta.Node) bool
	Populate(r Binder, d *meta.Node) error
}
