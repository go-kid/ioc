package ioc

import (
	"context"
	"flag"

	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/debug"
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
	if debug.HasRunDebugFlag() {
		return RunDebugWithContext(ctx, ops...)
	}
	return doRun(ctx, ops, nil)
}

func RunDebug(ops ...app.SettingOption) (*app.App, error) {
	return RunDebugWithContext(context.Background(), ops...)
}

func RunDebugWithContext(ctx context.Context, ops ...app.SettingOption) (*app.App, error) {
	debugOps, err := debug.Setup()
	if err != nil {
		return nil, err
	}
	return doRun(ctx, ops, debugOps)
}

func doRun(ctx context.Context, ops []app.SettingOption, extra []app.SettingOption) (*app.App, error) {
	s := app.NewApp()
	if flagLogLevel != "" {
		syslog.Level(syslog.NewLvFromString(flagLogLevel))
	}
	allOps := append(ops, registerHandlers...)
	registerHandlers = nil
	allOps = append(allOps, extra...)
	if err := s.RunWithContext(ctx, allOps...); err != nil {
		return nil, err
	}
	return s, nil
}
