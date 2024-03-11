package app

import (
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/samber/lo"
)

type SettingOption func(s *App)

func Options(opts ...SettingOption) SettingOption {
	return func(s *App) {
		for _, opt := range opts {
			opt(s)
		}
	}
}

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
	return SetScanPolicies(lo.Map(tags, func(item string, _ int) scanner.ScanPolicy {
		return scanner.DefaultScanPolicy(item, nil)
	})...)
}

func SetScanPolicies(policies ...scanner.ScanPolicy) SettingOption {
	return func(s *App) {
		s.Scanner.AddTags(policies)
	}
}

func AddCustomizedInjectors(injectors ...injector.InjectProcessor) SettingOption {
	return func(s *App) {
		s.Injector.AddCustomizedInjectors(injectors...)
	}
}

func LogLevel(lv syslog.Lv) SettingOption {
	return func(s *App) {
		syslog.Level(lv)
	}
}

func SetLogger(l syslog.Logger) SettingOption {
	return func(s *App) {
		syslog.SetLogger(l)
	}
}
