package ioc

import (
	. "github.com/go-kid/ioc/app"
)

func Run(ops ...SettingOption) error {
	s := NewApp(append([]SettingOption{
		SetRegistry(_registry),
	}, ops...)...)
	return s.Run()
}
