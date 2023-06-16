package configure

import (
	"fmt"
	"github.com/go-kid/ioc/meta"
	"github.com/spf13/viper"
	"reflect"
)

type DefaultBinder struct {
	v *viper.Viper
}

func (d *DefaultBinder) SetConfig(c any) error {
	v, ok := c.(*viper.Viper)
	if !ok {
		return fmt.Errorf("cfg is not *viper.Viper")
	}
	d.v = v
	return nil
}

func (d *DefaultBinder) PropInject(m *meta.Meta) error {
	unmarshall := func(key string, a interface{}) error {
		if key == "" {
			return d.v.Unmarshal(a)
		}
		return d.v.UnmarshalKey(key, a)
	}
	for _, prop := range m.Properties {
		var fieldType = prop.Type
		var isPtrType = false
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
			isPtrType = true
		}
		var val = reflect.New(fieldType)
		err := unmarshall(prop.Prefix, val.Interface())
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

func (n *NopBinder) SetConfig(c any) error {
	return nil
}

func (n *NopBinder) PropInject(m *meta.Meta) error {
	return nil
}
