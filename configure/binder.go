package configure

import (
	"bytes"
	"github.com/go-kid/ioc/defination"
	"github.com/spf13/viper"
	"reflect"
)

type ViperBinder struct {
	ct string
	v  *viper.Viper
}

func NewViperBinder(configType string) *ViperBinder {
	if configType == "" {
		configType = "yaml"
	}
	return &ViperBinder{
		ct: configType,
		v:  viper.New(),
	}
}

func (d *ViperBinder) SetConfig(c []byte) error {
	d.v.SetConfigType(d.ct)
	err := d.v.ReadConfig(bytes.NewBuffer(c))
	if err != nil {
		return err
	}
	return nil
}

func (d *ViperBinder) PropInject(properties []*defination.Node) error {
	unmarshall := func(key string, a interface{}) error {
		if key == "" {
			return d.v.Unmarshal(a)
		}
		return d.v.UnmarshalKey(key, a)
	}
	for _, prop := range properties {
		var fieldType = prop.Type
		var isPtrType = false
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
			isPtrType = true
		}
		var val = reflect.New(fieldType)
		err := unmarshall(prop.Tag, val.Interface())
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

type NopBinder struct{}

func (n *NopBinder) SetConfig(c []byte) error {
	return nil
}

func (n *NopBinder) PropInject(properties []*defination.Node) error {
	return nil
}
