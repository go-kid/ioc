package app

import (
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/syslog"
)

type SettingOption func(s *App)

func Options(ops ...SettingOption) SettingOption {
	return func(s *App) {
		for _, op := range ops {
			op(s)
		}
	}
}

func SetRegistry(r container.SingletonRegistry) SettingOption {
	return func(s *App) {
		s.registry = r
	}
}

func SetComponents(cs ...any) SettingOption {
	return func(s *App) {
		for _, c := range cs {
			s.registry.RegisterSingleton(c)
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

func SetFactory(factory container.Factory) SettingOption {
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
	LogTrace = LogLevel(syslog.LvTrace)
	LogDebug = LogLevel(syslog.LvDebug)
	LogWarn  = LogLevel(syslog.LvWarn)
	LogError = LogLevel(syslog.LvError)
)
