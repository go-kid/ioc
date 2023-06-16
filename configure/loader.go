package configure

import (
	"bytes"
	"github.com/spf13/viper"
	"os"
)

type DefaultLoader struct{}

func (c *DefaultLoader) LoadConfig(u string) (any, error) {
	v := viper.New()
	v.AddConfigPath(u)
	v.SetConfigType("yaml")
	data, err := os.ReadFile(u)
	if err != nil {
		return nil, err
	}

	err = v.ReadConfig(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return v, nil
}

type NopLoader struct{}

func (n *NopLoader) LoadConfig(u string) (any, error) { return nil, nil }
