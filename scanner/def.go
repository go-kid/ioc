package scanner

import (
	"github.com/go-kid/ioc/scanner/meta"
	"reflect"
)

type Scanner interface {
	AddTags(policies []ScanPolicy)
	ScanComponent(c any) *meta.Meta
	ScanNodes(source *meta.Holder, policy ScanPolicy) []*meta.Node
}

type ScanPolicy interface {
	Tag() string
	ExtHandler() ExtTagHandler
}

type ExtTagHandler func(field reflect.StructField, value reflect.Value) (tag string, tagVal string, err bool)
