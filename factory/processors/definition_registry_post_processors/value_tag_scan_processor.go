package definition_registry_post_processors

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/strings2"
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
					var argstr string
					i := strings2.IndexSkipBlocks(tagVal, ",")
					if i != -1 {
						tagVal, argstr = tagVal[:i], tagVal[i:]
					}
					tagVal = fmt.Sprintf("${%s}%s", tagVal, argstr)
				}
				return
			},
			Required: true,
		},
	}
}
