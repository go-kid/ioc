package definition_registry_post_processors

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
)

type configurationScanProcessor struct {
	processors.DefaultTagScanDefinitionRegistryPostProcessor
}

func NewConfigurationScanProcessor() processors.DefinitionRegistryPostProcessor {
	return &configurationScanProcessor{processors.DefaultTagScanDefinitionRegistryPostProcessor{
		NodeType: component_definition.PropertyTypeConfiguration,
		Tag:      definition.PrefixTag,
		ExtractHandler: func(meta *component_definition.Meta, field *component_definition.Field) (tag, tagVal string, ok bool) {
			if configuration, infer := field.Value.Interface().(definition.ConfigurationProperties); infer {
				tagVal = configuration.Prefix()
				ok = true
			}
			return
		},
	}}
}
