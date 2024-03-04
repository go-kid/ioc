package app

import (
	"errors"
	"fmt"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"sort"
	"sync"
)

type App struct {
	configLoaders []configure.Loader
	configure.Binder
	injector.Injector
	registry.Registry
	factory.Factory
	scanner.Scanner
	configPath              string
	enableApplicationRunner bool
}

func NewApp(ops ...SettingOption) *App {
	var s = &App{
		configLoaders:           []configure.Loader{&loader.FileLoader{}}, //default use file loader
		Binder:                  binder.NewViperBinder("yaml"),            //default use viper binder and 'yaml' config type
		Injector:                injector.Default(),
		Registry:                registry.GlobalRegistry(),
		Factory:                 factory.Default(),
		Scanner:                 scanner.Default(),
		configPath:              "",
		enableApplicationRunner: true,
	}
	for _, op := range ops {
		op(s)
	}
	err := s.validate()
	if err != nil {
		syslog.Fatal(err)
	}
	return s
}

func (s *App) validate() error {
	if s.configPath != "" {
		if len(s.configLoaders) == 0 {
			return errors.New("missing config loader")
		}
		if s.Binder == nil {
			return errors.New("missing config binder")
		}
	}
	if s.Injector == nil {
		return errors.New("missing injector")
	}
	if s.Registry == nil {
		return errors.New("missing registry")
	}
	if s.Factory == nil {
		return errors.New("missing factory")
	}
	if s.Scanner == nil {
		return errors.New("missing scanner")
	}
	return nil
}

func (s *App) Run() error {
	/* begin scan component to meta */
	syslog.Info("start scanning registered component...")
	s.Registry.Scan(s.Scanner)
	syslog.Info("component scanning finished, registry ready")
	/* registry ready */

	/* begin load and bind configuration */
	syslog.Info("start init config...")
	if err := s.initConfig(); err != nil {
		return fmt.Errorf("init config failed: %v", err)
	}
	syslog.Info("configuration ready")
	/* configuration ready */

	/* set default init behavior */
	s.Factory.SetIfNilPostInitFunc(defaultPostInitFunc(s.Registry))
	syslog.Info("factory ready")
	/* factory ready */

	/* begin inject dependencies */
	syslog.Info("start wire dependencies...")
	if err := s.wire(); err != nil {
		return fmt.Errorf("wire dependencies failed: %v", err)
	}
	syslog.Info("dependencies injection finished, components ready")
	/* dependency injection ready */

	/* begin call runners */
	syslog.Info("start start runners...")
	if err := s.callRunners(); err != nil {
		return fmt.Errorf("start runners failed: %v", err)
	}
	syslog.Info("all runners started")
	/* finished */
	return nil
}

func (s *App) initConfig() error {
	if s.configPath == "" {
		syslog.Trace("config path is empty, skip init configs")
		return nil
	}
	syslog.Tracef("using config path %s init configs...", s.configPath)

	for i, l := range s.configLoaders {
		syslog.Tracef("config loaders start loading config %s ...[%d/%d]", reflectx.Id(l), i+1, len(s.configLoaders))
		config, err := l.LoadConfig(s.configPath)
		if err != nil {
			return fmt.Errorf("config loader load config failed: %v", err)
		}
		syslog.Tracef("config loader loading finished ...[%d/%d]", i+1, len(s.configLoaders))
		err = s.Binder.SetConfig(config)
		if err != nil {
			return fmt.Errorf("config binder set config failed: %v", err)
		}
	}

	metas := s.GetComponents()
	for _, m := range metas {
		err := s.Binder.PropInject(m.Properties)
		if err != nil {
			return fmt.Errorf("binder inject prop failed: %v", err)
		}
	}
	return nil
}

func getComponentPostProcessors(r registry.Registry) []defination.ComponentPostProcessor {
	postMetas := r.GetComponents(registry.Interface(new(defination.ComponentPostProcessor)))
	postProcessors := make([]defination.ComponentPostProcessor, 0, len(postMetas))
	for _, pm := range postMetas {
		postProcessors = append(postProcessors, pm.Raw.(defination.ComponentPostProcessor))
		syslog.Tracef("collecting post processors %s", pm.ID())
		r.RemoveComponents(pm.Name)
	}
	return postProcessors
}

func (s *App) wire() error {
	components := s.GetComponents()
	//Reduce recursion depth
	syslog.Trace("sorting components")
	sort.Slice(components, func(i, j int) bool {
		if len(components[i].DependsBy) != len(components[j].DependsBy) {
			return len(components[i].DependsBy) > len(components[j].DependsBy)
		}
		return len(components[i].AllDependencies()) < len(components[j].AllDependencies())
	})
	for _, m := range components {
		err := s.Initialize(s.Registry, s.Injector, m)
		if err != nil {
			return fmt.Errorf("initialize component failed: %v", err)
		}
	}
	return nil
}

func (s *App) callRunners() error {
	if !s.enableApplicationRunner {
		return nil
	}
	metas := s.GetComponents(registry.Interface(new(defination.ApplicationRunner)))
	var runners = lo.Map(metas, func(item *meta.Meta, _ int) defination.ApplicationRunner {
		return item.Raw.(defination.ApplicationRunner)
	})
	if len(runners) == 0 {
		syslog.Trace("find 0 runner(s), skip")
		return nil
	}
	syslog.Tracef("find %d runner(s), start sort", len(runners))
	sort.Slice(runners, func(i, j int) bool {
		return runners[i].Order() < runners[j].Order()
	})
	for i := range runners {
		runner := runners[i]
		syslog.Tracef("start runner %s [%d/%d]", reflectx.Id(runner), i+1, len(runners))
		err := runner.Run()
		if err != nil {
			return fmt.Errorf("start runner %s failed: %v", reflectx.Id(runner), err)
		}
	}
	return nil
}

func (s *App) Close() {
	metas := s.GetComponents(registry.Interface(new(defination.CloserComponent)))
	wg := sync.WaitGroup{}
	wg.Add(len(metas))
	for _, m := range metas {
		go func(m *meta.Meta) {
			defer wg.Done()
			if err := m.Raw.(defination.CloserComponent).Close(); err != nil {
				syslog.Errorf("Error closing %s", m.ID())
			} else {
				syslog.Infof("close component: %s", m.ID())
			}
		}(m)
	}
	wg.Wait()
}

func defaultPostInitFunc(r registry.Registry) factory.MetaFunc {
	postMetas := r.GetComponents(registry.Interface(new(defination.ComponentPostProcessor)))
	postProcessors := make([]defination.ComponentPostProcessor, 0, len(postMetas))
	for _, pm := range postMetas {
		postProcessors = append(postProcessors, pm.Raw.(defination.ComponentPostProcessor))
		syslog.Tracef("collecting post processors %s", pm.ID())
		r.RemoveComponents(pm.Name)
	}
	return func(m *meta.Meta) error {
		// before process
		for _, processor := range postProcessors {
			err := processor.PostProcessBeforeInitialization(m.Raw)
			if err != nil {
				return fmt.Errorf("post processor: %T process before %s init error: %v", processor, m.ID(), err)
			}
		}
		// init
		if ic, ok := m.Raw.(defination.InitializeComponent); ok {
			syslog.Trace("component %s is InitializeComponent, start init", m.ID())
			err := ic.Init()
			if err != nil {
				return fmt.Errorf("component %s inited failed: %s", m.ID(), err)
			}
		}

		// after process
		for _, processor := range postProcessors {
			err := processor.PostProcessAfterInitialization(m.Raw)
			if err != nil {
				return fmt.Errorf("post processor: %T process after %s init error: %v", processor, m.ID(), err)
			}
		}
		return nil
	}
}
