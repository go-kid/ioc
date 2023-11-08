package loader

import "github.com/go-kid/ioc/configure"

type RawLoader struct{}

func NewRawLoader() configure.Loader {
	return &RawLoader{}
}

func (r *RawLoader) LoadConfig(u string) ([]byte, error) {
	return []byte(u), nil
}
