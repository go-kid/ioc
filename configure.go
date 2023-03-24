package ioc

import (
	"bytes"
	"github.com/spf13/viper"
	"os"
	"reflect"
)

var conf *viper.Viper

func initConfig(cfgPath string) error {
	v, err := loadConfigWithFile(cfgPath)
	if err != nil {
		return err
	}
	conf = v
	return nil
}

func loadConfigWithFile(fileName string) (*viper.Viper, error) {
	c := viper.New()
	c.AddConfigPath(fileName)
	c.SetConfigType("yaml")
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	err = c.ReadConfig(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return c, nil
}

func unmarshall(key string, a interface{}) error {
	if key == "" {
		return conf.Unmarshal(a)
	}
	return conf.UnmarshalKey(key, a)
}

func propInject(m *meta) error {
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
