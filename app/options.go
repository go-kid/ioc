package app

import (
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner"
	"github.com/go-kid/ioc/scanner/meta"
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

func SetConfigLoader(loaders ...configure.Loader) SettingOption {
	return func(s *App) {
		s.configLoaders = loaders
	}
}

func AppendConfigLoader(loaders ...configure.Loader) SettingOption {
	return func(s *App) {
		s.configLoaders = append(s.configLoaders, loaders...)
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

func DisableComponentInitialization() SettingOption {
	return func(s *App) {
		s.Factory.SetIfNilPostInitFunc(func(m *meta.Meta) error {
			return nil
		})
	}
}

func SetScanner(sc scanner.Scanner) SettingOption {
	return func(s *App) {
		s.Scanner = sc
	}
}

func SetScanTags(tags ...string) SettingOption {
	return func(s *App) {
		s.Scanner.AddTags(tags)
	}
}

func AddCustomizedInjectors(injectors ...injector.InjectProcessor) SettingOption {
	return func(s *App) {
		s.Injector.AddCustomizedInjectors(injectors...)
	}
}
