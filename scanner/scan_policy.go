package scanner

import (
	"github.com/go-kid/ioc/scanner/meta"
)

type policy struct {
	nt      meta.NodeType
	tag     string
	handler ExtTagHandler
}

func (p *policy) Group() meta.NodeType {
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
		nt:      meta.NodeTypeComponent,
		tag:     tag,
		handler: handler,
	}
}

func NewConfigurationScanPolicy(tag string, handler ExtTagHandler) ScanPolicy {
	return &policy{
		nt:      meta.NodeTypeConfiguration,
		tag:     tag,
		handler: handler,
	}
}
