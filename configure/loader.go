package configure

import (
	"os"
)

type DefaultLoader struct{}

func (c *DefaultLoader) LoadConfig(u string) ([]byte, error) {
	return os.ReadFile(u)
}

type NopLoader struct{}

func (n *NopLoader) LoadConfig(u string) ([]byte, error) { return nil, nil }

type RawLoader struct{}

func (r *RawLoader) LoadConfig(u string) ([]byte, error) {
	return []byte(u), nil
}
