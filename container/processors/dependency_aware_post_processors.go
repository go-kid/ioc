package processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
	"reflect"
)

type dependencyAwarePostProcessors struct {
	DefaultTagScanDefinitionRegistryPostProcessor
	DefaultInstantiationAwareComponentPostProcessor
	Registry container.DefinitionRegistry
}

func NewDependencyAwarePostProcessors() container.InstantiationAwareComponentPostProcessor {
	return &dependencyAwarePostProcessors{
		DefaultTagScanDefinitionRegistryPostProcessor: DefaultTagScanDefinitionRegistryPostProcessor{
			NodeType:       component_definition.PropertyTypeComponent,
			Tag:            definition.InjectTag,
			ExtractHandler: nil,
			Required:       true,
		},
	}
}

func (d *dependencyAwarePostProcessors) PostProcessComponentFactory(factory container.Factory) error {
	d.Registry = factory.GetDefinitionRegistry()
	return nil
}

func (d *dependencyAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (d *dependencyAwarePostProcessors) Order() int {
	return OrderDependencyAware
}

func (d *dependencyAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if prop.Tag != definition.InjectTag {
			continue
		}
		//aware by type
		if prop.TagVal == "" {
			typeOption := getMetaTypeOption(prop.Type)
			if typeOption == nil {
				continue
			}
			dm := d.Registry.GetMetas(typeOption)
			prop.Injects = append(prop.Injects, dm...)
			continue
		}
		//aware by name
		if prop.TagVal != "" && (prop.Type.Kind() == reflect.Ptr || prop.Type.Kind() == reflect.Interface) {
			dm := d.Registry.GetMetaByName(prop.TagVal)
			prop.Injects = append(prop.Injects, dm)
		}
	}
	return nil, nil
}

func getMetaTypeOption(typ reflect.Type) container.Option {
	var typeOption container.Option
	if p, ok := isActualKind(typ, reflect.Pointer); ok {
		typeOption = container.Type(p)
	} else if p, ok = isActualKind(typ, reflect.Interface); ok {
		typeOption = container.InterfaceType(p)
	}
	return typeOption
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
