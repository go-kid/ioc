package processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
	"reflect"
)

type dependencyTypeAwarePostProcessors struct {
	DefaultInstantiationAwareComponentPostProcessor
	definition.LazyInitComponent
	Registry container.DefinitionRegistry
}

func NewDependencyTypeAwarePostProcessors() container.InstantiationAwareComponentPostProcessor {
	return &dependencyTypeAwarePostProcessors{}
}

func (d *dependencyTypeAwarePostProcessors) PostProcessComponentFactory(factory container.Factory) error {
	d.Registry = factory.GetDefinitionRegistry()
	return nil
}

func (d *dependencyTypeAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (d *dependencyTypeAwarePostProcessors) Order() int {
	return OrderDependencyAware
}

func (d *dependencyTypeAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if prop.Tag != definition.InjectTag {
			continue
		}
		if prop.TagVal == "" {
			var typeOption container.Option
			if p, ok := isActualKind(prop.Type, reflect.Pointer); ok {
				typeOption = container.Type(p)
			} else if p, ok = isActualKind(prop.Type, reflect.Interface); ok {
				typeOption = container.InterfaceType(p)
			} else {
				continue
			}
			dm := d.Registry.GetMetas(typeOption)
			prop.Injects = append(prop.Injects, dm...)
		}
	}
	return nil, nil
}
