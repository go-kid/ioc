package processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

type loggerAwarePostProcessors struct {
	DefaultTagScanDefinitionRegistryPostProcessor
	DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
}

func NewLoggerAwarePostProcessor() container.InstantiationAwareComponentPostProcessor {
	return &loggerAwarePostProcessors{
		DefaultTagScanDefinitionRegistryPostProcessor: DefaultTagScanDefinitionRegistryPostProcessor{
			LazyInitComponent: definition.LazyInitComponent{},
			NodeType:          "Logger",
			Tag:               definition.LoggerTag,
			ExtractHandler:    nil,
			Required:          true,
		},
	}
}

func (d *loggerAwarePostProcessors) Order() int {
	return PriorityOrderLoggerAware
}

func (d *loggerAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (d *loggerAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, property := range properties {
		if property.Tag == definition.LoggerTag && reflectx.IsTypeImplement(property.Type, new(syslog.Logger)) {
			var pref = property.TagStr
			if pref == "" {
				if property.Args().Has("embed") {
					pref = property.Holder.String()
				} else {
					pref = property.Holder.Meta.String()
				}
			}
			logger := syslog.Pref(pref)
			property.Value.Set(reflect.ValueOf(logger))
		}
	}
	return properties, nil
}
