package meta

import (
	"github.com/samber/lo"
	"strings"
)

type Source struct {
	*Base
	Meta    *Meta
	IsEmbed bool
	Source  *Source
}

func (s *Source) ID() string {
	if s.IsEmbed {
		var ids []string
		_ = s.Walk(func(source *Source) error {
			if source.IsEmbed {
				ids = append(ids, source.Type.Name())
			} else {
				ids = append(ids, source.Meta.ID())
			}
			return nil
		})
		ids = lo.Reverse(ids)
		return strings.Join(ids, ".")
		//return s.Embeds.ID()
	}
	return s.Meta.ID()
}

func (s *Source) Walk(f func(source *Source) error) error {
	if s == nil {
		return nil
	}
	var n = s
	for n != nil {
		err := f(n)
		if err != nil {
			return nil
		}
		n = n.Source
	}
	return nil
}
