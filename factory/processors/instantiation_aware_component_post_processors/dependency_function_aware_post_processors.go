package instantiation_aware_component_post_processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/factory/support"
	"reflect"
)

type dependencyFunctionAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	Registry support.DefinitionRegistry
}

func NewDependencyFunctionAwarePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &dependencyFunctionAwarePostProcessors{}
}

func (d *dependencyFunctionAwarePostProcessors) PostProcessComponentFactory(factory factory.Factory) error {
	d.Registry = factory.GetDefinitionRegistry()
	return nil
}

func (d *dependencyFunctionAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (d *dependencyFunctionAwarePostProcessors) Order() int {
	return 100
}

func (d *dependencyFunctionAwarePostProcessors) PostProcessProperties(properties []*component_definition.Node, component any, componentName string) ([]*component_definition.Node, error) {
	for _, prop := range properties {
		if prop.Tag != definition.FuncTag {
			continue
		}
		var (
			typeOption support.Option
			funcOption support.Option
		)
		if p, ok := isActualKind(prop.Type, reflect.Pointer); ok {
			typeOption = support.Type(p)
		} else if p, ok = isActualKind(prop.Type, reflect.Interface); ok {
			typeOption = support.InterfaceType(p)
		} else {
			continue
		}
		if args, ok := prop.Args().Find("returns"); ok {
			fmt.Println(prop.TagVal, "args", args)
			var options []support.Option
			for _, arg := range args {
				options = append(options, support.FuncNameAndResult(prop.TagVal, arg))
			}
			funcOption = support.Or(options...)
		} else {
			fmt.Println(prop.TagVal)
			funcOption = support.FuncName(prop.TagVal)
		}
		dm := d.Registry.GetMetas(typeOption, funcOption)
		prop.Injects = append(prop.Injects, dm...)
	}
	return nil, nil
}
