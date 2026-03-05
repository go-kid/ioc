package debug

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"runtime"

	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/container/factory"
	"github.com/go-kid/ioc/syslog"
)

type hookSetter interface {
	SetFactoryHook(hook container.FactoryHook)
}

type contextSetter interface {
	SetContext(ctx context.Context)
}

type DebugOption func(*DebugFactory)

func WithDryRun() DebugOption {
	return func(df *DebugFactory) {
		df.dryRun = true
	}
}

type DebugFactory struct {
	inner      container.Factory
	controller *Controller
	server     *Server
	collector  *Collector
	hook       *debugHook
	dryRun     bool
}

func newDebugFactory(opts ...DebugOption) *DebugFactory {
	inner := factory.Default()
	ctrl := NewController()
	coll := NewCollector()

	staticSub, err := fs.Sub(StaticFS, "static")
	if err != nil {
		panic(fmt.Sprintf("debug static fs: %v", err))
	}
	srv := NewServer(ctrl, coll, staticSub)
	hook := newHook(ctrl, srv, coll)

	if hs, ok := inner.(hookSetter); ok {
		hs.SetFactoryHook(hook)
	}

	df := &DebugFactory{
		inner:      inner,
		controller: ctrl,
		server:     srv,
		collector:  coll,
		hook:       hook,
	}
	for _, opt := range opts {
		opt(df)
	}
	srv.dryRun = df.dryRun
	return df
}

// Setup creates the debug infrastructure, starts the debug server,
// opens the browser, and returns the app.SettingOption slice to apply.
// It also auto-detects the --ioc:dry_run CLI flag.
func Setup(opts ...DebugOption) ([]app.SettingOption, error) {
	if hasCLIFlag("--ioc:dry_run") {
		opts = append(opts, WithDryRun())
	}

	df := newDebugFactory(opts...)
	addr, err := df.server.Start()
	if err != nil {
		return nil, fmt.Errorf("debug server start: %w", err)
	}

	url := "http://" + addr
	syslog.Pref("Debug").Infof("debug server started at %s", url)
	if df.dryRun {
		syslog.Pref("Debug").Infof("dry run mode enabled, component initialization and runners will be skipped")
	}
	openBrowser(url)

	appOpts := []app.SettingOption{app.SetFactory(df)}
	if df.dryRun {
		appOpts = append(appOpts, app.SkipRunners())
	}
	return appOpts, nil
}

// HasCLIFlag checks whether --ioc:run_debug is present in os.Args.
func HasRunDebugFlag() bool {
	return hasCLIFlag("--ioc:run_debug")
}

func hasCLIFlag(flag string) bool {
	for _, arg := range os.Args[1:] {
		if arg == flag {
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

func (df *DebugFactory) Close() {
	df.controller.Close()
}

// --- container.Factory delegation ---

func (df *DebugFactory) GetRegisteredComponents() map[string]any {
	return df.inner.GetRegisteredComponents()
}

func (df *DebugFactory) GetDefinitionRegistryPostProcessors() []container.DefinitionRegistryPostProcessor {
	return df.inner.GetDefinitionRegistryPostProcessors()
}

func (df *DebugFactory) SetRegistry(r container.SingletonRegistry) {
	if df.dryRun {
		r.RegisterSingleton(&dryRunPostProcessor{})
	}
	df.inner.SetRegistry(r)
}

func (df *DebugFactory) SetConfigure(c configure.Configure) {
	df.inner.SetConfigure(c)
}

func (df *DebugFactory) PrepareComponents() error {
	return df.inner.PrepareComponents()
}

func (df *DebugFactory) Refresh() error {
	return df.inner.Refresh()
}

func (df *DebugFactory) GetComponents(opts ...container.Option) ([]any, error) {
	return df.inner.GetComponents(opts...)
}

func (df *DebugFactory) GetComponentByName(name string) (any, error) {
	return df.inner.GetComponentByName(name)
}

func (df *DebugFactory) GetConfigure() configure.Configure {
	return df.inner.GetConfigure()
}

func (df *DebugFactory) GetDefinitionRegistry() container.DefinitionRegistry {
	return df.inner.GetDefinitionRegistry()
}

func (df *DebugFactory) SetContext(ctx context.Context) {
	if cs, ok := df.inner.(contextSetter); ok {
		cs.SetContext(ctx)
	}
}
