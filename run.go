package ioc

import (
	"context"
	"flag"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/syslog"
)

var (
	flagLogLevel     string
	registerHandlers []app.SettingOption
)

func init() {
	flag.StringVar(&flagLogLevel, "logLevel", "", "set ioc app log level")
}

func Register(cs ...interface{}) {
	registerHandlers = append(registerHandlers, app.SetComponents(cs...))
}

func Run(ops ...app.SettingOption) (*app.App, error) {
	return RunWithContext(context.Background(), ops...)
}

func RunWithContext(ctx context.Context, ops ...app.SettingOption) (*app.App, error) {
	s := app.NewApp()
	if flagLogLevel != "" {
		syslog.Level(syslog.NewLvFromString(flagLogLevel))
	}
	allOps := append(ops, registerHandlers...)
	registerHandlers = nil
	if err := s.RunWithContext(ctx, allOps...); err != nil {
		return nil, err
	}
	return s, nil
}
