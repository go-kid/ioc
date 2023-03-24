package configure

import (
	"bytes"
	"fmt"
	"github.com/kid-hash/kid-ioc/meta"
	"github.com/spf13/viper"
	"reflect"
)

type Configure struct {
	v          *viper.Viper
	Config     []byte
	ConfigType string
}

func NewConfigure() *Configure {
	return &Configure{
		v: viper.New(),
	}
}

func (c *Configure) Load() error {
	if c.Config == nil || c.ConfigType == "" {
		return nil
	}
	c.v.SetConfigType(c.ConfigType)
	err := c.v.ReadConfig(bytes.NewBuffer(c.Config))
	if err != nil {
		return fmt.Errorf("viper read config failed: %v\nconfig:%s\nconfig type:%s", err, c.Config, c.ConfigType)
	}
	return nil
}

func (c *Configure) unmarshall(key string, a interface{}) error {
	if key == "" {
		return c.v.Unmarshal(a)
	}
	return c.v.UnmarshalKey(key, a)
}

func (c *Configure) PropInject(m *meta.Meta) error {
	for _, prop := range m.Properties {
		var fieldType = prop.Type
		var isPtrType = false
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
			isPtrType = true
		}
		var val = reflect.New(fieldType)
		err := c.unmarshall(prop.Prefix, val.Interface())
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

//var conf *viper.Viper
//
//func InitConfig(cfgPath string) error {
//	if cfgPath == "" {
//		return nil
//	}
//
//	switch true {
//	case strings.HasSuffix(cfgPath, ".yaml"), strings.HasSuffix(cfgPath, ".yml"):
//		v, err := loadByFile(cfgPath)
//		if err != nil {
//			return err
//		}
//		conf = v
//	}
//
//	return nil
//}
//
//func loadByFile(fileName string) (*viper.Viper, error) {
//	c := viper.New()
//	c.AddConfigPath(fileName)
//	c.SetConfigType("yaml")
//	data, err := os.ReadFile(fileName)
//	if err != nil {
//		return nil, err
//	}
//
//	err = c.ReadConfig(bytes.NewBuffer(data))
//	if err != nil {
//		return nil, err
//	}
//	return c, nil
//}
