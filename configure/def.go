package configure

import (
	"github.com/go-kid/ioc/scanner/meta"
)

type Loader interface {
	LoadConfig(u string) ([]byte, error)
}

type Binder interface {
	SetConfig(c []byte) error
	CompareWith(newConfig []byte, path string) (bool, error)
	PropInject(properties []*meta.Node) error
}
