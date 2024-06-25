package processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/util/framework_helper"
	"reflect"
)

type constructorAwarePostProcessors struct {
	DefaultTagScanDefinitionRegistryPostProcessor
	DefaultInstantiationAwareComponentPostProcessor
	Registry              container.DefinitionRegistry
	componentConstructors map[string]any
}

func NewConstructorAwarePostProcessors() container.InstantiationAwareComponentPostProcessor {
	return &constructorAwarePostProcessors{
		DefaultTagScanDefinitionRegistryPostProcessor: DefaultTagScanDefinitionRegistryPostProcessor{
			NodeType:       component_definition.PropertyTypeComponent,
			Tag:            definition.InjectTag,
			ExtractHandler: nil,
			Required:       true,
		},
		componentConstructors: map[string]any{},
	}
}

func (d *constructorAwarePostProcessors) PostProcessComponentFactory(factory container.Factory) error {
	d.Registry = factory.GetDefinitionRegistry()
	return nil
}

func (d *constructorAwarePostProcessors) PostProcessDefinitionRegistry(registry container.DefinitionRegistry, component any, componentName string) error {
	var fn = reflect.TypeOf(component)
	if fn.Kind() != reflect.Func || fn.NumOut() != 1 {
		return nil
	}

	var method = reflect.ValueOf(component)
	var inValuesAddr []reflect.Value
	for i := 0; i < fn.NumIn(); i++ {
		var val reflect.Value
		switch in := fn.In(i); in.Kind() {
		case reflect.Pointer:
			val = reflect.New(in).Elem()
			val.Set(reflect.New(in.Elem()))
		}
		inValuesAddr = append(inValuesAddr, val)
	}
	var inValues = inValuesAddr
	//var inValues = make([]reflect.Value, len(inValuesAddr))
	//for i, addr := range inValuesAddr {
	//	inValues[i] = addr.Elem()
	//}
	var results []reflect.Value
	if fn.IsVariadic() {
		results = method.CallSlice(inValues)
	} else {
		results = method.Call(inValues)
	}
	instance := results[0].Interface()

	instanceName := framework_helper.GetComponentName(instance)
	meta := registry.GetMetaOrRegister(instanceName, instance)

	var properties []*component_definition.Property
	for i := 0; i < fn.NumIn(); i++ {
		in := fn.In(i)
		val := inValuesAddr[i]
		property := component_definition.NewProperty(
			&component_definition.Field{
				Base: &component_definition.Base{
					Type:  in,
					Value: val,
				},
				Holder:      component_definition.NewHolder(meta),
				StructField: reflect.StructField{},
			},
			component_definition.PropertyTypeComponent,
			"wire", "",
		)
		properties = append(properties, property)
		fmt.Println(property)
	}
	meta.SetProperties(properties...)

	return d.DefaultTagScanDefinitionRegistryPostProcessor.PostProcessDefinitionRegistry(registry, instance, instanceName)
}

func (d *constructorAwarePostProcessors) Order() int {
	return OrderConstructorAware
}

type A struct {
	Name string
}

func (a *A) GetName() string {
	return a.Name
}

type IA interface {
	GetName() string
}
