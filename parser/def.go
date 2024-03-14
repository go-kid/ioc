package parser

import (
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/registry"
)

type Parser interface {
	Parse(b configure.Binder, r registry.Registry) error
	AddStrategies(s ...ParseStrategy)
}

type ParseStrategy interface {
	Parse(b configure.Binder, r registry.Registry, exp string) (string, bool)
}
