package app

import (
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
	s.validate()
	return s
}

func (s *App) validate() {
	if s.configPath != "" {
		if len(s.configLoaders) == 0 {
			syslog.Panic("missing config loader")
		}
		if s.Binder == nil {
			syslog.Panic("missing config binder")
		}
	}
	if s.Injector == nil {
		syslog.Panic("missing injector")
	}
	if s.Registry == nil {
		syslog.Panic("missing registry")
	}
	if s.Factory == nil {
		syslog.Panic("missing factory")
	}
	if s.Scanner == nil {
		syslog.Panic("missing scanner")
	}
}

func (s *App) Run() error {
	/* begin scan component to meta */
	s.Registry.Scan(s.Scanner)
	/* registry ready */

	/* set default init behavior */
	s.Factory.SetIfNilPostInitFunc(defaultPostInitFunc(getComponentPostProcessors(s.Registry)))
	/* factory ready */

	/* begin load and bind configuration */
	if err := s.initConfig(); err != nil {
		return err
	}
	/* configuration ready */

	/* begin inject dependencies */
	if err := s.wire(); err != nil {
		return err
	}
	/* dependency injection ready */

	/* begin call runners */
	if err := s.callRunners(); err != nil {
		return fmt.Errorf("runners failed: %v", err)
	}
	/* finished */
	return nil
}

func (s *App) initConfig() error {
	if s.configPath == "" {
		return nil
	}

	for _, l := range s.configLoaders {
		config, err := l.LoadConfig(s.configPath)
		if err != nil {
			return fmt.Errorf("load config failed: %v", err)
		}
		err = s.Binder.SetConfig(config)
		if err != nil {
			return fmt.Errorf("init config failed: %v", err)
		}
	}

	metas := s.GetComponents()
	for _, m := range metas {
		err := s.Binder.PropInject(m.Properties)
		if err != nil {
			return err
		}
	}
	return nil
}

func getComponentPostProcessors(r registry.Registry) []defination.ComponentPostProcessor {
	postMetas := r.GetComponents(registry.Interface(new(defination.ComponentPostProcessor)))
	postProcessors := make([]defination.ComponentPostProcessor, 0, len(postMetas))
	for _, pm := range postMetas {
		postProcessors = append(postProcessors, pm.Raw.(defination.ComponentPostProcessor))
		r.RemoveComponents(pm.Name)
	}
	return postProcessors
}

func (s *App) wire() error {
	components := s.GetComponents()
	//Reduce recursion depth
	sort.Slice(components, func(i, j int) bool {
		if len(components[i].DependsBy) != len(components[j].DependsBy) {
			return len(components[i].DependsBy) > len(components[j].DependsBy)
		}
		return len(components[i].AllDependencies()) < len(components[j].AllDependencies())
	})
	for _, m := range components {
		err := s.Initialize(s.Registry, s.Injector, m)
		if err != nil {
			return fmt.Errorf("initialize failed: %v", err)
		}
	}
	return nil
}

func defaultPostInitFunc(postProcessors []defination.ComponentPostProcessor) factory.MetaFunc {
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
			err := ic.Init()
			if err != nil {
				return fmt.Errorf("component: %s inited failed: %s", m.ID(), err)
			}
		}
		syslog.Infof("initialize component: %s\n", m.ID())

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

func (s *App) callRunners() error {
	if !s.enableApplicationRunner {
		return nil
	}
	metas := s.GetComponents(registry.Interface(new(defination.ApplicationRunner)))
	var runners = lo.Map(metas, func(item *meta.Meta, _ int) defination.ApplicationRunner {
		return item.Raw.(defination.ApplicationRunner)
	})
	sort.Slice(runners, func(i, j int) bool {
		return runners[i].Order() < runners[j].Order()
	})
	for i := range runners {
		err := runners[i].Run()
		if err != nil {
			return err
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
				syslog.Infof("Error closing %s", m.ID())
			} else {
				syslog.Infof("close component: %s", m.ID())
			}
		}(m)
	}
	wg.Wait()
}
