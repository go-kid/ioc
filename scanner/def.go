package scanner

import (
	"github.com/go-kid/ioc/scanner/meta"
	"reflect"
)

type Scanner interface {
	ScanComponent(c any) *meta.Meta
	ScanNodes(tag string, t reflect.Type, v reflect.Value, handlers ...ExtTagHandler) []*meta.Node
}
