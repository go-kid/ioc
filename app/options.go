package app

import (
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner"
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

func SetConfigLoader(loader configure.Loader) SettingOption {
	return func(s *App) {
		s.Loader = loader
	}
}

func SetConfigBinder(binder configure.Binder) SettingOption {
	return func(s *App) {
		s.Binder = binder
	}
}

func DisableApplicationRunner() SettingOption {
	return func(s *App) {
		s.enableApplicationRunner = false
	}
}

func SetScanner(sc scanner.Scanner) SettingOption {
	return func(s *App) {
		s.Registry.SetScanner(sc)
	}
}
