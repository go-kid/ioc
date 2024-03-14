package parser

import (
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/registry"
)

type parser struct {
	strategies []ParseStrategy
}

func Default() Parser {
	return &parser{}
}

func (p *parser) AddStrategies(s ...ParseStrategy) {
	p.strategies = append(p.strategies, s...)
}

func (p *parser) Parse(b configure.Binder, r registry.Registry) error {
	//TODO implement me
	panic("implement me")
}
