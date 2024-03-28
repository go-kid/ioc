package scanner

import (
	"github.com/go-kid/ioc/component_definition"
	"reflect"
)

type Scanner interface {
	AddScanPolicies(policies ...ScanPolicy)
	ScanComponent(c any) *component_definition.Meta
}

type ScanPolicy interface {
	Group() component_definition.NodeType
	Tag() string
	ExtHandler() ExtTagHandler
}

type ExtTagHandler func(field reflect.StructField, value reflect.Value) (tag string, tagVal string, ok bool)
