package loader

import (
	"os"
)

type FileLoader struct{}

func (c *FileLoader) LoadConfig(u string) ([]byte, error) {
	return os.ReadFile(u)
}
