package configure

import (
	"github.com/go-kid/ioc/injector"
)

type Loader interface {
	LoadConfig(u string) ([]byte, error)
}

type Binder interface {
	SetConfig(c []byte) error
	PropInject(properties []*injector.Node) error
}
