package processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
	"reflect"
)

type dependencyFunctionAwarePostProcessors struct {
	DefaultTagScanDefinitionRegistryPostProcessor
	DefaultInstantiationAwareComponentPostProcessor
	Registry container.DefinitionRegistry
}

func NewDependencyFunctionAwarePostProcessors() container.InstantiationAwareComponentPostProcessor {
	return &dependencyFunctionAwarePostProcessors{
		DefaultTagScanDefinitionRegistryPostProcessor: DefaultTagScanDefinitionRegistryPostProcessor{
			NodeType:       component_definition.PropertyTypeComponent,
			Tag:            definition.FuncTag,
			ExtractHandler: nil,
			Required:       true,
		},
	}
}

func (d *dependencyFunctionAwarePostProcessors) PostProcessComponentFactory(factory container.Factory) error {
	d.Registry = factory.GetDefinitionRegistry()
	return nil
}

func (d *dependencyFunctionAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (d *dependencyFunctionAwarePostProcessors) Order() int {
	return OrderDependencyAware
}

func (d *dependencyFunctionAwarePostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if prop.Tag != definition.FuncTag {
			continue
		}
		var (
			typeOption container.Option
			funcOption container.Option
		)
		if p, ok := isActualKind(prop.Type, reflect.Pointer); ok {
			typeOption = container.Type(p)
		} else if p, ok = isActualKind(prop.Type, reflect.Interface); ok {
			typeOption = container.InterfaceType(p)
		} else {
			continue
		}
		if args, ok := prop.Args().Find("returns"); ok {
			var options []container.Option
			for _, arg := range args {
				options = append(options, container.FuncNameAndResult(prop.TagVal, arg))
			}
			funcOption = container.Or(options...)
		} else {
			funcOption = container.FuncName(prop.TagVal)
		}
		dm := d.Registry.GetMetas(typeOption, funcOption)
		prop.Injects = append(prop.Injects, dm...)
	}
	return nil, nil
}
