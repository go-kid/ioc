package definition_registry_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"reflect"
)

// PropTagScanProcessor @DefinitionRegistryPostProcessor
type PropTagScanProcessor struct {
}

func (d *PropTagScanProcessor) PostProcessDefinitionRegistry(registry factory.DefinitionRegistry, component any, componentName string) error {
	meta := registry.GetMetaByName(componentName)
	if meta == nil {
		meta = component_definition.NewMeta(component)
		registry.RegisterMeta(meta)
	}
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
