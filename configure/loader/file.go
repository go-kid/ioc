package loader

import (
	"github.com/pkg/errors"
	"os"
)

type FileLoader string

func (c FileLoader) Priority() {
}

func (c FileLoader) Order() int {
	return 0
}

func NewFileLoader(file string) FileLoader {
	return FileLoader(file)
}

func (c FileLoader) LoadConfig() ([]byte, error) {
	bytes, err := os.ReadFile(string(c))
	if err != nil {
		return nil, errors.Wrapf(err, "read file: %s", string(c))
	}
	return bytes, nil
}
