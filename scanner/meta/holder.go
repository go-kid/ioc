package meta

import (
	"fmt"
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
		return fmt.Sprintf("%s.Embed(%s)", s.Holder.ID(), s.Type.Name())
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
