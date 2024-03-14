package scanner

import (
	"github.com/go-kid/ioc/scanner/meta"
	"reflect"
)

type Scanner interface {
	AddScanPolicies(policies ...ScanPolicy)
	ScanComponent(c any) *meta.Meta
}

type ScanPolicy interface {
	Group() meta.NodeType
	Tag() string
	ExtHandler() ExtTagHandler
}

type ExtTagHandler func(field reflect.StructField, value reflect.Value) (tag string, tagVal string, ok bool)
