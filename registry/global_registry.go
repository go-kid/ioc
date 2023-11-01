package registry

var _registry = NewRegistry()

func GlobalRegistry() Registry {
	return _registry
}
