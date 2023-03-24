package kid_ioc

import (
	"encoding/json"
	"github.com/kid-hash/kid-ioc/configure"
	"github.com/kid-hash/kid-ioc/factory"
	"github.com/kid-hash/kid-ioc/registry"
	"os"
	"path/filepath"
)

type setting struct {
	registry  *registry.Registry
	configure *configure.Configure
}

type SettingOption func(s *setting)

func Run(ops ...SettingOption) error {
	var s = &setting{
		registry:  _registry,
		configure: configure.NewConfigure(),
	}
	for _, op := range ops {
		op(s)
	}
	err := s.configure.Load()
	if err != nil {
		return err
	}
	f := factory.NewFactory(s.registry, s.configure)
	err = f.Start()
	if err != nil {
		return err
	}
	return nil
}

func SetComponents(c ...interface{}) SettingOption {
	return func(s *setting) {
		s.registry.Register(c...)
	}
}

func SetConfigStructure(v any) SettingOption {
	return func(s *setting) {
		bytes, err := json.Marshal(v)
		if err != nil {
			panic("marshall config failed: " + err.Error())
		}
		s.configure.Config = bytes
		s.configure.ConfigType = "json"
	}
}

func SetConfigPath(path string) SettingOption {
	return func(s *setting) {
		ext := filepath.Ext(path)
		s.configure.ConfigType = ext[1:]
		bytes, err := os.ReadFile(path)
		if err != nil {
			panic("read config failed: " + err.Error())
		}
		s.configure.Config = bytes
	}
}

func SetConfigSrc(src []byte, configType string) SettingOption {
	return func(s *setting) {
		s.configure.Config = src
		s.configure.ConfigType = configType
	}
}
