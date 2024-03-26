package factory

var _registry = NewRegistry()

func GlobalRegistry() SingletonRegistry {
	return _registry
}
