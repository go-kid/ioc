package processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
)

type DefaultTagScanDefinitionRegistryPostProcessor struct {
	definition.LazyInitComponent
	NodeType       component_definition.PropertyType
	Tag            string
	ExtractHandler func(meta *component_definition.Meta, field *component_definition.Field) (tag, tagVal string, ok bool)
	Required       bool
}

func (d *DefaultTagScanDefinitionRegistryPostProcessor) PostProcessDefinitionRegistry(registry container.DefinitionRegistry, component any, componentName string) error {
	meta := registry.GetMetaOrRegister(componentName, component)
	var properties []*component_definition.Property
	for _, field := range meta.Fields {
		if d.Tag != "" {
			if tagVal, ok := field.StructField.Tag.Lookup(d.Tag); ok {
				properties = append(properties, component_definition.NewProperty(field, d.NodeType, d.Tag, tagVal))
				continue
			}
		}

		if d.ExtractHandler != nil {
			if tag, tagVal, ok := d.ExtractHandler(meta, field); ok {
				if tag == "" {
					tag = d.Tag
				}
				properties = append(properties, component_definition.NewProperty(field, d.NodeType, tag, tagVal))
			}
		}
	}

	for _, property := range properties {
		if d.Required && !property.Args().Has(component_definition.ArgRequired) {
			property.SetArg(component_definition.ArgRequired)
		}
	}

	meta.SetProperties(properties...)
	return nil
}
