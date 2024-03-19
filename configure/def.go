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
	PropInject(properties []*meta.Node) error
}

type Configure interface {
	Binder
	AddLoaders(loaders ...Loader)
	SetLoaders(loaders ...Loader)
	SetBinder(binder Binder)
	Initialize(metas ...*meta.Meta) error
	Populate(metas ...*meta.Meta) error
}
