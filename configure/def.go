package configure

import (
	"github.com/go-kid/ioc/defination"
)

type ConfigLoader interface {
	LoadConfig(u string) ([]byte, error)
}

type ConfigBinder interface {
	SetConfig(c []byte) error
	PropInject(properties []*defination.Node) error
}
