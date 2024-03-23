package binder

import (
	"bytes"
	"github.com/spf13/viper"
)

type ViperBinder struct {
	configType string
	Viper      *viper.Viper
}

func NewViperBinder(configType string) *ViperBinder {
	if configType == "" {
		configType = "yaml"
	}
	v := viper.New()
	v.SetConfigType(configType)
	return &ViperBinder{
		configType: configType,
		Viper:      v,
	}
}

func (d *ViperBinder) SetConfig(c []byte) error {
	err := d.Viper.MergeConfig(bytes.NewBuffer(c))
	if err != nil {
		return err
	}
	return nil
}

func (d *ViperBinder) Get(path string) any {
	if path == "" {
		return d.Viper.AllSettings()
	}
	return d.Viper.Get(path)
}

func (d *ViperBinder) Set(path string, val any) {
	d.Viper.Set(path, val)
}

func (d *ViperBinder) Unmarshall(key string, a any) error {
	if key == "" {
		return d.Viper.Unmarshal(a)
	}
	return d.Viper.UnmarshalKey(key, a)
}
