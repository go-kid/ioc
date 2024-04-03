package instantiation_aware_component_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/factory/support"
	"reflect"
)

type dependencyNameAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.LazyInitComponent
	Registry support.DefinitionRegistry
}

func NewDependencyNameAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &dependencyNameAwarePostProcessors{}
}

func (d *dependencyNameAwarePostProcessors) PostProcessComponentFactory(factory factory.Factory) error {
	d.Registry = factory.GetDefinitionRegistry()
	return nil
}

func (d *dependencyNameAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (d *dependencyNameAwarePostProcessors) Order() int {
	return OrderDependencyAware
}

func (d *dependencyNameAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if prop.Tag != definition.InjectTag {
			continue
		}
		if prop.TagVal != "" && (prop.Type.Kind() == reflect.Ptr || prop.Type.Kind() == reflect.Interface) {
			dm := d.Registry.GetMetaByName(prop.TagVal)
			prop.Injects = append(prop.Injects, dm)
		}
	}
	return nil, nil
}
