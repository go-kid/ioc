package ioc

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"

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
	if hasRunDebugFlag() {
		return RunDebug(ops...)
	}
	return RunWithContext(context.Background(), ops...)
}

func RunWithContext(ctx context.Context, ops ...app.SettingOption) (*app.App, error) {
	if hasRunDebugFlag() {
		return RunDebugWithContext(ctx, ops...)
	}
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

func RunDebug(ops ...app.SettingOption) (*app.App, error) {
	return RunDebugWithContext(context.Background(), ops...)
}

func RunDebugWithContext(ctx context.Context, ops ...app.SettingOption) (*app.App, error) {
	s := app.NewApp()
	if flagLogLevel != "" {
		syslog.Level(syslog.NewLvFromString(flagLogLevel))
	}

	df := debug.NewDebugFactory()
	addr, err := df.StartServer()
	if err != nil {
		return nil, fmt.Errorf("debug server start: %w", err)
	}

	url := "http://" + addr
	syslog.Pref("Debug").Infof("debug server started at %s", url)
	openBrowser(url)

	allOps := append(ops, registerHandlers...)
	registerHandlers = nil
	allOps = append(allOps, app.SetFactory(df))

	if err := s.RunWithContext(ctx, allOps...); err != nil {
		return nil, err
	}
	return s, nil
}

func hasRunDebugFlag() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--ioc:run_debug" {
			return true
		}
	}
	return false
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	_ = cmd.Start()
}
