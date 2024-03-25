package scanner

import (
	"github.com/go-kid/ioc/component_definition"
)

type policy struct {
	nt      component_definition.NodeType
	tag     string
	handler ExtTagHandler
}

func (p *policy) Group() component_definition.NodeType {
	return p.nt
}

func (p *policy) Tag() string {
	return p.tag
}

func (p *policy) ExtHandler() ExtTagHandler {
	return p.handler
}

func NewComponentScanPolicy(tag string, handler ExtTagHandler) ScanPolicy {
	return &policy{
		nt:      component_definition.NodeTypeComponent,
		tag:     tag,
		handler: handler,
	}
}

func NewConfigurationScanPolicy(tag string, handler ExtTagHandler) ScanPolicy {
	return &policy{
		nt:      component_definition.NodeTypeConfiguration,
		tag:     tag,
		handler: handler,
	}
}
