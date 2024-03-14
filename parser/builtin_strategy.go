package parser

import (
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/registry"
	"regexp"
)

type configureStrategy struct {
}

func (c *configureStrategy) Parse(b configure.Binder, r registry.Registry, exp string) (string, bool) {

}

type componentStrategy struct {
}

func newComponentStrategy() ParseStrategy {
	regexp.MustCompile("(\\#\\{[\\d\\w]*\\.[\\d\\w]+(:[\\d\\w]*)?\\})+")
	return &componentStrategy{}
}

func (c *componentStrategy) Parse(b configure.Binder, r registry.Registry, exp string) (string, bool) {
	//TODO implement me
	panic("implement me")
}
