package definition_registry_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/support"
	"reflect"
)

// PropTagScanProcessor @DefinitionRegistryPostProcessor
type PropTagScanProcessor struct {
}

func (d *PropTagScanProcessor) PostProcessDefinitionRegistry(registry support.DefinitionRegistry, component any, componentName string) error {
	meta := registry.GetMetaOrRegister(componentName, func() *component_definition.Meta {
		return component_definition.NewMeta(component)
	})
	var (
		nodeType = component_definition.NodeTypeConfiguration
		tag      = definition.PropTag
	)
	var nodes []*component_definition.Node
	for _, field := range meta.Fields {
		if tagVal, ok := field.StructField.Tag.Lookup(tag); ok {
			nodes = append(nodes, component_definition.NewNode(field, nodeType, tag, tagVal))
			continue
		}
		if tagVal, ok := propHandler(field.Value); ok {
			nodes = append(nodes, component_definition.NewNode(field, nodeType, tag, tagVal))
		}
	}
	meta.SetNodes(nodes...)
	return nil
}

func propHandler(value reflect.Value) (tagVal string, ok bool) {
	if configuration, infer := value.Interface().(definition.Configuration); infer {
		tagVal = configuration.Prefix()
		ok = true
	}
	return
}
