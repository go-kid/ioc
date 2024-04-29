package configure

type Loader interface {
	LoadConfig() ([]byte, error)
}

type Binder interface {
	SetConfig(c []byte) error
	Get(path string) any
	Set(path string, val any)
}

type Configure interface {
	Binder
	AddLoaders(loaders ...Loader)
	SetLoaders(loaders ...Loader)
	SetBinder(binder Binder)
	Initialize() error
}
