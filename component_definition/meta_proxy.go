package component_definition

type MetaProxy struct {
	Parent *Meta
	*Meta
}

func ProxyFor(parent *Meta, proxy any) *MetaProxy {
	return &MetaProxy{
		Parent: parent,
		Meta:   proxy.(*Meta),
	}
}
