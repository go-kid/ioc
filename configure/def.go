package configure

import (
	"github.com/go-kid/ioc/scanner/meta"
)

type Loader interface {
	LoadConfig(u string) ([]byte, error)
}

type Binder interface {
	SetConfig(c []byte) error
	Get(path string) any
	PropInject(properties []*meta.Node) error
}
