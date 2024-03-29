package app

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/support"
	"github.com/go-kid/ioc/syslog"
)

type SettingOption func(s *App)

func SetRegistry(r support.SingletonRegistry) SettingOption {
	return func(s *App) {
		s.registry = r
	}
}

func SetComponents(cs ...any) SettingOption {
	return func(s *App) {
		for _, c := range cs {
			name, alias := component_definition.GetComponentNameWithAlias(c)
			if alias != "" {
				name = alias
			}
			s.registry.RegisterSingleton(name, c)
		}
	}
}

func SetConfigure(c configure.Configure) SettingOption {
	return func(s *App) {
		s.Configure = c
	}
}

func SetConfig(cfg string) SettingOption {
	return func(s *App) {
		s.Configure.AddLoaders(loader.NewFileLoader(cfg))
	}
}

func SetFactory(factory factory.Factory) SettingOption {
	return func(s *App) {
		s.Factory = factory
	}
}

func SetConfigLoader(loaders ...configure.Loader) SettingOption {
	return func(s *App) {
		s.Configure.SetLoaders(loaders...)
	}
}

func AddConfigLoader(loaders ...configure.Loader) SettingOption {
	return func(s *App) {
		s.Configure.SetLoaders(loaders...)
	}
}

func SetConfigBinder(binder configure.Binder) SettingOption {
	return func(s *App) {
		s.Configure.SetBinder(binder)
	}
}

func DisableApplicationRunner() SettingOption {
	return func(s *App) {
		s.enableApplicationRunner = false
	}
}

//func SetScanner(sc scanner.Scanner) SettingOption {
//	return func(s *App) {
//		s.Scanner = sc
//	}
//}

//func AddScanPolicies(policies ...scanner.ScanPolicy) SettingOption {
//	return func(s *App) {
//		s.Scanner.AddScanPolicies(policies...)
//	}
//}

func AddPopulateProcessors(processors ...configure.PopulateProcessor) SettingOption {
	return func(s *App) {
		s.Configure.AddPopulateProcessors(processors...)
	}
}

func AddInjectionRules(rules ...factory.InjectionRule) SettingOption {
	return func(s *App) {
		s.Factory.AddInjectionRules(rules...)
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

var (
	LogTrace = func(s *App) { syslog.Level(syslog.LvTrace) }
	LogWarn  = func(s *App) { syslog.Level(syslog.LvWarn) }
	LogError = func(s *App) { syslog.Level(syslog.LvError) }
)
