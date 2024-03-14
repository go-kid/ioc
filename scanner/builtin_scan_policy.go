package scanner

import (
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/scanner/meta"
	"reflect"
)

type scanComponentPolicy struct {
}

func (s *scanComponentPolicy) Group() meta.NodeType {
	return meta.NodeTypeComponent
}

func (s *scanComponentPolicy) Tag() string {
	return defination.InjectTag
}

func (s *scanComponentPolicy) ExtHandler() ExtTagHandler {
	return nil
}

type scanConfigurationPolicy struct {
}

func (s *scanConfigurationPolicy) Group() meta.NodeType {
	return meta.NodeTypeConfiguration
}

func (s *scanConfigurationPolicy) Tag() string {
	return defination.PropTag
}

func (s *scanConfigurationPolicy) ExtHandler() ExtTagHandler {
	return func(field reflect.StructField, value reflect.Value) (tag string, tagVal string, ok bool) {
		if configuration, infer := value.Interface().(defination.Configuration); infer {
			tag = defination.PropTag
			tagVal = configuration.Prefix()
			ok = true
		}
		return
	}
}
