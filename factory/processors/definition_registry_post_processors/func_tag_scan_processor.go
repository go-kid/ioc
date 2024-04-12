package definition_registry_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
)

type funcTagScanProcessor struct {
	processors.DefaultTagScanDefinitionRegistryPostProcessor
}

func NewFuncTagScanProcessor() processors.DefinitionRegistryPostProcessor {
	return &funcTagScanProcessor{
		processors.DefaultTagScanDefinitionRegistryPostProcessor{
			NodeType:       component_definition.PropertyTypeComponent,
			Tag:            definition.FuncTag,
			ExtractHandler: nil,
			Required:       true,
		},
	}
}
