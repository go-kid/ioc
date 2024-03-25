package registry

var _registry = NewRegistry()

func GlobalRegistry() SingletonRegistry {
	return _registry
}
