package instantiation_aware_component_post_processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/util/fas"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

type dependencyFurtherMatchingPostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.LazyInitComponent
}

func NewDependencyFurtherMatchingProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &dependencyFurtherMatchingPostProcessors{}
}

func (d *dependencyFurtherMatchingPostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (d *dependencyFurtherMatchingPostProcessors) Order() int {
	return OrderDependencyFurtherMatching
}

func (d *dependencyFurtherMatchingPostProcessors) PostProcessProperties(properties []*component_definition.Property, component any, componentName string) ([]*component_definition.Property, error) {
	for _, prop := range properties {
		if prop.PropertyType != component_definition.PropertyTypeComponent {
			continue
		}
		dependencies, err := filterDependencies(prop, prop.Injects)
		if err != nil {
			if len(dependencies) == 0 {
				if prop.IsRequired() {
					return nil, fmt.Errorf("field '%s' is required but not found any components\n caused by: %v", prop.String(), err)
				}
				return nil, nil
			}
			return nil, err
		}
		prop.Injects = dependencies
	}
	return nil, nil
}

var (
	primaryInterface = new(definition.WirePrimary)
)

func filterDependencies(n *component_definition.Property, metas []*component_definition.Meta) ([]*component_definition.Meta, error) {
	//remove nil meta
	result := fas.Filter(metas, func(m *component_definition.Meta) bool {
		return m != nil
	})
	if len(result) == 0 {
		return nil, fmt.Errorf("%s not found available components", n.ID())
	}
	//filter qualifier
	if qualifierName, isQualifier := n.Args().Find(component_definition.ArgQualifier); isQualifier {
		result = fas.Filter(result, func(m *component_definition.Meta) bool {
			qualifier, ok := m.Raw.(definition.WireQualifier)
			return ok && n.Args().Has(component_definition.ArgQualifier, qualifier.Qualifier())
		})
		if len(result) == 0 {
			return nil, fmt.Errorf("no component found for qualifier %s", qualifierName)
		}
	}

	//filter primary for single type
	if len(result) > 1 && n.Type.Kind() != reflect.Slice && n.Type.Kind() != reflect.Array {
		var candidate = result[0]

		for _, m := range result {
			//Primary interface first
			if reflectx.IsTypeImplement(m.Type, primaryInterface) {
				candidate = m
				break
			}
			//non naming component is preferred in multiple candidates
			if !m.IsAlias() {
				candidate = m
			}
		}
		result = []*component_definition.Meta{candidate}
	}
	return result, nil
}
