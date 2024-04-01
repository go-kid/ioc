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

type dependencyValidatePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
}

func NewDependencyValidatePostProcessors() processors.InstantiationAwareComponentPostProcessor {
	return &dependencyValidatePostProcessors{}
}

func (d *dependencyValidatePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (d *dependencyValidatePostProcessors) Order() int {
	return 102
}

func (d *dependencyValidatePostProcessors) PostProcessProperties(properties []*component_definition.Node, component any, componentName string) ([]*component_definition.Node, error) {
	for _, prop := range properties {
		if prop.NodeType != component_definition.NodeTypeComponent {
			continue
		}
		dependencies, err := filterDependencies(prop, prop.Injects)
		if err != nil {
			if len(dependencies) == 0 {
				if prop.Args().Has(component_definition.ArgRequired, "true") {
					return nil, fmt.Errorf("field '%s' is required but not found any components", prop.ID())
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

func filterDependencies(n *component_definition.Node, metas []*component_definition.Meta) ([]*component_definition.Meta, error) {
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
			return nil, fmt.Errorf("field %s: no component found for qualifier %s", n.ID(), qualifierName)
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
