package definition_registry_post_processors

import (
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
			NodeType:       component_definition.NodeTypeConfiguration,
			Tag:            definition.ValueTag,
			ExtractHandler: nil,
		},
	}
}
