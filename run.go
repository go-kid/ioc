package ioc

import (
	"flag"
	. "github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/syslog"
)

func Register(cs ...interface{}) {
	Settings(SetComponents(cs...))
}

var (
	flagLogLevel string
)

func init() {
	flag.StringVar(&flagLogLevel, "logLevel", "", "set ioc app log level")
}

func Run(ops ...SettingOption) (*App, error) {
	s := NewApp()
	if flagLogLevel != "" {
		syslog.Level(syslog.NewLvFromString(flagLogLevel))
	}
	if err := s.Run(ops...); err != nil {
		return nil, err
	}
	return s, nil
}
