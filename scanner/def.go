package scanner

import (
	"github.com/go-kid/ioc/scanner/meta"
)

type Scanner interface {
	AddTags(tags []string)
	ScanComponent(c any) *meta.Meta
	ScanNodes(source *meta.Source, tag string, handlers ...ExtTagHandler) []*meta.Node
}
