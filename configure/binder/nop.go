package binder

import (
	"github.com/go-kid/ioc/scanner/meta"
)

type NopBinder struct{}

func (n *NopBinder) SetConfig(c []byte) error {
	return nil
}

func (n *NopBinder) PropInject(properties []*meta.Node) error {
	return nil
}
