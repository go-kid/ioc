package configure

import (
	"github.com/go-kid/ioc/meta"
)

type ConfigLoader interface {
	LoadConfig(u string) (any, error)
}

type ConfigBinder interface {
	SetConfig(c any) error
	PropInject(m *meta.Meta) error
}
