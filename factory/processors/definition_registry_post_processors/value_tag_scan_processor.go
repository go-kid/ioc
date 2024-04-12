package definition_registry_post_processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
)

type valueTagScanProcessor struct {
	processors.DefaultTagScanDefinitionRegistryPostProcessor
}

func NewValueTagScanProcessor() processors.DefinitionRegistryPostProcessor {
	return &valueTagScanProcessor{
		processors.DefaultTagScanDefinitionRegistryPostProcessor{
			NodeType: component_definition.PropertyTypeConfiguration,
			Tag:      definition.ValueTag,
			ExtractHandler: func(meta *component_definition.Meta, field *component_definition.Field) (tag, tagVal string, ok bool) {
				if tagVal, ok = field.StructField.Tag.Lookup(definition.PropTag); ok {
					tagVal = fmt.Sprintf("${%s}", tagVal)
				}
				return
			},
		},
	}
}
