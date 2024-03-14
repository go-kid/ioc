package loader

import (
	"os"
)

type FileLoader string

func NewFileLoader(file string) FileLoader {
	return FileLoader(file)
}

func (c FileLoader) LoadConfig() ([]byte, error) {
	return os.ReadFile(string(c))
}
