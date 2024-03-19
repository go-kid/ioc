package binder

import (
	"bytes"
	"fmt"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/spf13/viper"
	"reflect"
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
	return d.Viper.Get(path)
}

func (d *ViperBinder) Set(path string, val any) {
	d.Viper.Set(path, val)
}

func (d *ViperBinder) PropInject(properties []*meta.Node) error {
	for _, prop := range properties {
		syslog.Tracef("viper binder start bind config %s, prefix: %s", prop.ID(), prop.TagVal)
		var fieldType = prop.Type
		var isPtrType = false
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
			isPtrType = true
		}
		var val = reflect.New(fieldType)
		err := d.unmarshall(prop.TagVal, val.Interface())
		if err != nil {
			return fmt.Errorf("viper binder bind config %s, prefix: %s error: %v", prop.ID(), prop.TagVal, err)
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
