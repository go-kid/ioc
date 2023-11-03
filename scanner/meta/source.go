package meta

import "reflect"

type Source struct {
	Source      *Source
	IsAnonymous bool
	Type        reflect.Type
	Value       reflect.Value
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
