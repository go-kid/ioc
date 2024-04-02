package ioc

import (
	. "github.com/go-kid/ioc/app"
)

func Register(cs ...interface{}) {
	Settings(SetComponents(cs...))
}

func Run(ops ...SettingOption) (*App, error) {
	s := NewApp(ops...)
	if err := s.Run(); err != nil {
		return nil, err
	}
	return s, nil
}
