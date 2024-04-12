package definition_registry_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
)

type wireTagScanProcessor struct {
	processors.DefaultTagScanDefinitionRegistryPostProcessor
}

func NewWireTagScanProcessor() processors.DefinitionRegistryPostProcessor {
	return &wireTagScanProcessor{
		processors.DefaultTagScanDefinitionRegistryPostProcessor{
			NodeType:       component_definition.PropertyTypeComponent,
			Tag:            definition.InjectTag,
			ExtractHandler: nil,
			Required:       true,
		},
	}
}
