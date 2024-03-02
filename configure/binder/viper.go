package binder

import (
	"bytes"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/spf13/viper"
	"reflect"
)

type ViperBinder struct {
	configType string
	Viper      *viper.Viper
}

func NewViperBinder(configType string) configure.Binder {
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
	return d.Viper.Get(path)
}

func (d *ViperBinder) PropInject(properties []*meta.Node) error {
	for _, prop := range properties {
		var fieldType = prop.Type
		var isPtrType = false
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
			isPtrType = true
		}
		var val = reflect.New(fieldType)
		err := d.unmarshall(prop.TagVal, val.Interface())
		if err != nil {
			return err
		}
		if isPtrType {
			prop.Value.Set(val)
		} else {
			prop.Value.Set(val.Elem())
		}
	}
	return nil
}

func (d *ViperBinder) unmarshall(key string, a interface{}) error {
	if key == "" {
		return d.Viper.Unmarshal(a)
	}
	return d.Viper.UnmarshalKey(key, a)
}
