package ioc

import (
	. "github.com/go-kid/ioc/app"
)

func Run(ops ...SettingOption) error {
	s := NewApp(append([]SettingOption{
		SetRegistry(_registry),
		SetDefaultConfigure(),
	}, ops...)...)
	return s.Run()
}
