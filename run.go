package ioc

import (
	. "github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure"
)

func Run(ops ...SettingOption) error {
	s := NewApp(append([]SettingOption{
		SetRegistry(_registry),
		SetConfigure(&configure.DefaultLoader{}, &configure.DefaultBinder{}),
	}, ops...)...)
	return s.Run()
}
