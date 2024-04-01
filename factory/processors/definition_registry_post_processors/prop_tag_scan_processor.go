package definition_registry_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
)

type propTagScanProcessor struct {
	processors.DefaultTagScanDefinitionRegistryPostProcessor
}

func NewPropTagScanProcessor() processors.DefinitionRegistryPostProcessor {
	return &propTagScanProcessor{processors.DefaultTagScanDefinitionRegistryPostProcessor{
		NodeType: component_definition.NodeTypeConfiguration,
		Tag:      definition.PropTag,
		ExtractHandler: func(meta *component_definition.Meta, field *component_definition.Field) (tagVal string, ok bool) {
			if configuration, infer := field.Value.Interface().(definition.Configuration); infer {
				tagVal = configuration.Prefix()
				ok = true
			}
			return
		},
	}}
}
