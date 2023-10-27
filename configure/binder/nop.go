package binder

import "github.com/go-kid/ioc/injector"

type NopBinder struct{}

func (n *NopBinder) SetConfig(c []byte) error {
	return nil
}

func (n *NopBinder) PropInject(properties []*injector.Node) error {
	return nil
}
