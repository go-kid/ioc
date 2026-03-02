package component_definition

import (
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/util/reflectx"
)

var wirePrimaryInterface = new(definition.WirePrimary)

// SelectBestCandidate picks the best match from multiple Meta candidates
// using the priority: WirePrimary > non-alias > first element.
func SelectBestCandidate(metas []*Meta) *Meta {
	if len(metas) == 1 {
		return metas[0]
	}
	candidate := metas[0]
	for _, m := range metas {
		if reflectx.IsTypeImplement(m.Type, wirePrimaryInterface) {
			return m
		}
		if !m.IsAlias() {
			candidate = m
		}
	}
	return candidate
}
