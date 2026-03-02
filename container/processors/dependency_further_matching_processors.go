package processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
	"github.com/pkg/errors"
	"reflect"
)

type dependencyFurtherMatchingPostProcessors struct {
	DefaultInstantiationAwareComponentPostProcessor
	definition.LazyInitComponent
}

func NewDependencyFurtherMatchingProcessors() container.InstantiationAwareComponentPostProcessor {
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
					return nil, errors.WithMessagef(err, "field '%s' is required but not found any components", prop.String())
				}
				return nil, nil
			}
			return nil, err
		}
		prop.Injects = dependencies
	}
	return nil, nil
}

func filterDependencies(n *component_definition.Property, metas []*component_definition.Meta) ([]*component_definition.Meta, error) {
	var result []*component_definition.Meta
	for _, m := range metas {
		if m != nil {
			result = append(result, m)
		}
	}
	if len(result) == 0 {
		return nil, errors.Errorf("inject '%s' not found available components", n)
	}
	if qualifierName, isQualifier := n.Args().Find(component_definition.ArgQualifier); isQualifier {
		var qualified []*component_definition.Meta
		for _, m := range result {
			qualifier, ok := m.Raw.(definition.WireQualifier)
			if ok && n.Args().Has(component_definition.ArgQualifier, qualifier.Qualifier()) {
				qualified = append(qualified, m)
			}
		}
		result = qualified
		if len(result) == 0 {
			return nil, errors.Errorf("inject '%s' matching qualifier '%s' not found available components", n, qualifierName)
		}
	}

	if len(result) > 1 && n.Type.Kind() != reflect.Slice && n.Type.Kind() != reflect.Array {
		result = []*component_definition.Meta{component_definition.SelectBestCandidate(result)}
	}
	return result, nil
}
