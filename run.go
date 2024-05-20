package ioc

import (
	"flag"
	. "github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/syslog"
)

var (
	flagLogLevel     string
	registerHandlers []SettingOption
)

func init() {
	flag.StringVar(&flagLogLevel, "logLevel", "", "set ioc app log level")
}

func Register(cs ...interface{}) {
	registerHandlers = append(registerHandlers, SetComponents(cs...))
}

func Run(ops ...SettingOption) (*App, error) {
	s := NewApp()
	if flagLogLevel != "" {
		syslog.Level(syslog.NewLvFromString(flagLogLevel))
	}
	if err := s.Run(append(ops, registerHandlers...)...); err != nil {
		return nil, err
	}
	return s, nil
}
