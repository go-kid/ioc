package factory

import "github.com/go-kid/ioc/factory/support"

var _registry = support.NewRegistry()

func GlobalRegistry() support.SingletonRegistry {
	return _registry
}
