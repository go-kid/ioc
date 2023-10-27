package app

import (
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/registry"
)

type SettingOption func(s *App)

func SetRegistry(r registry.Registry) SettingOption {
	return func(s *App) {
		s.Registry = r
	}
}

func SetComponents(c ...interface{}) SettingOption {
	return func(s *App) {
		s.Registry.Register(c...)
	}
}

func SetConfig(cfg string) SettingOption {
	return func(s *App) {
		s.configPath = cfg
	}
}

func SetFactory(factory factory.Factory) SettingOption {
	return func(s *App) {
		s.Factory = factory
	}
}

func SetConfigure(loader configure.ConfigLoader, binder configure.ConfigBinder) SettingOption {
	return func(s *App) {
		s.ConfigLoader = loader
		s.ConfigBinder = binder
	}
}

func SetDefaultConfigure() SettingOption {
	return func(s *App) {
		s.ConfigLoader = &configure.DefaultLoader{}
		s.ConfigBinder = configure.NewViperBinder("")
	}
}

func SetCallRunnersFunc(fn func(runners []defination.ApplicationRunner) error) SettingOption {
	return func(s *App) {
		s.callRunnersFunc = fn
	}
}
