package ioc

import (
	. "github.com/go-kid/ioc/app"
)

func Run(ops ...SettingOption) (*App, error) {
	s := NewApp(append([]SettingOption{
		SetRegistry(_registry),
	}, ops...)...)
	return s, s.Run()
}
