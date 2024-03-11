package meta

import (
	"github.com/samber/lo"
	"strings"
)

type Holder struct {
	*Base
	Meta    *Meta
	IsEmbed bool
	Holder  *Holder
}

func NewHolder(m *Meta) *Holder {
	return &Holder{
		Base: m.Base,
		Meta: m,
	}
}

func NewEmbedHolder(base *Base, holder *Holder) *Holder {
	return &Holder{
		Base:    base,
		Meta:    holder.Meta,
		IsEmbed: true,
		Holder:  holder,
	}
}

func (s *Holder) ID() string {
	if s.IsEmbed {
		var ids []string
		_ = s.Walk(func(source *Holder) error {
			if source.IsEmbed {
				ids = append(ids, source.Type.Name())
			} else {
				ids = append(ids, source.Meta.ID())
			}
			return nil
		})
		ids = lo.Reverse(ids)
		return strings.Join(ids, ".")
	}
	return s.Meta.ID()
}

func (s *Holder) Walk(f func(source *Holder) error) error {
	if s == nil {
		return nil
	}
	var n = s
	for n != nil {
		err := f(n)
		if err != nil {
			return nil
		}
		n = n.Holder
	}
	return nil
}
