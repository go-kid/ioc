package processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/factory/support"
)

type DefaultTagScanDefinitionRegistryPostProcessor struct {
	NodeType       component_definition.NodeType
	Tag            string
	ExtractHandler func(meta *component_definition.Meta, field *component_definition.Field) (tagVal string, ok bool)
}

func (d *DefaultTagScanDefinitionRegistryPostProcessor) PostProcessDefinitionRegistry(registry support.DefinitionRegistry, component any, componentName string) error {
	meta := registry.GetMetaOrRegister(componentName, func() *component_definition.Meta {
		return component_definition.NewMeta(component)
	})
	var nodes []*component_definition.Node
	for _, field := range meta.Fields {
		if tagVal, ok := field.StructField.Tag.Lookup(d.Tag); ok {
			nodes = append(nodes, component_definition.NewNode(field, d.NodeType, d.Tag, tagVal))
			continue
		}
		if d.ExtractHandler != nil {
			if tagVal, ok := d.ExtractHandler(meta, field); ok {
				nodes = append(nodes, component_definition.NewNode(field, d.NodeType, d.Tag, tagVal))
			}
		}
	}

	meta.SetNodes(nodes...)
	return nil
}
