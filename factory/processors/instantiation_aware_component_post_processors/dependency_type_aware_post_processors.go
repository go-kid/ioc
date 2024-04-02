package instantiation_aware_component_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/factory/support"
	"reflect"
)

type dependencyTypeAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	Registry support.DefinitionRegistry
}

func NewDependencyTypeAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &dependencyTypeAwarePostProcessors{}
}

func (d *dependencyTypeAwarePostProcessors) PostProcessComponentFactory(factory factory.Factory) error {
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
			var typeOption support.Option
			if p, ok := isActualKind(prop.Type, reflect.Pointer); ok {
				typeOption = support.Type(p)
			} else if p, ok = isActualKind(prop.Type, reflect.Interface); ok {
				typeOption = support.InterfaceType(p)
			} else {
				continue
			}
			dm := d.Registry.GetMetas(typeOption)
			prop.Injects = append(prop.Injects, dm...)
		}
	}
	return nil, nil
}

func isActualKind(p reflect.Type, kind reflect.Kind) (reflect.Type, bool) {
	if p.Kind() == kind {
		return p, true
	}
	if p.Kind() == reflect.Slice && p.Elem().Kind() == kind {
		return p.Elem(), true
	}
	return nil, false
}
