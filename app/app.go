package app

import (
	"errors"
	"fmt"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/samber/lo"
	"sort"
)

type App struct {
	configLoaders []configure.Loader
	configure.Binder
	registry.Registry
	factory.Factory
	scanner.Scanner
	configPath              string
	postProcessors          []defination.ComponentPostProcessor
	enableApplicationRunner bool
}

func NewApp(ops ...SettingOption) *App {
	var s = &App{
		configLoaders:           []configure.Loader{&loader.FileLoader{}}, //default use file loader
		Binder:                  binder.NewViperBinder("yaml"),            //default use viper binder and 'yaml' config type
		Registry:                registry.GlobalRegistry(),
		Factory:                 &factory.DefaultFactory{},
		Scanner:                 nil,
		configPath:              "",
		postProcessors:          nil,
		enableApplicationRunner: true,
	}
	for _, op := range ops {
		op(s)
	}
	return s
}

func (s *App) Run() error {
	err := s.initRegistry()
	if err != nil {
		return err
	}
	/* registry ready */

	//init configure
	if err := s.initConfig(); err != nil {
		return err
	}

	/* config ready */

	//init factory
	s.SetIfNilPostInitFunc(s.defaultPostInitFunc)

	/* factory ready */

	s.initComponentPostProcessors()

	/* post processors ready */

	if err := s.wire(); err != nil {
		return err
	}

	/* components ready */

	if err := s.callRunners(); err != nil {
		return fmt.Errorf("runners failed: %v", err)
	}
	return nil
}

//func (s *App) initProduceComponents() {
//	metas := s.GetComponents()
//	produces := lo.FlatMap[*meta.Meta, *meta.Meta](metas, func(item *meta.Meta, _ int) []*meta.Meta {
//		return item.Produce
//	})
//	lo.ForEach(produces, func(item *meta.Meta, _ int) {
//		s.Register(item)
//	})
//}

func (s *App) initConfig() error {
	if s.configPath == "" {
		return nil
	}
	if len(s.configLoaders) < 1 {
		return errors.New("config loader is not available")
	}
	if s.Binder == nil {
		return errors.New("config binder is not available")
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

func (s *App) initComponentPostProcessors() {
	metas := s.GetComponents(registry.Interface(new(defination.ComponentPostProcessor)))
	s.postProcessors = make([]defination.ComponentPostProcessor, 0, len(metas))
	for _, m := range metas {
		s.postProcessors = append(s.postProcessors, m.Raw.(defination.ComponentPostProcessor))
		s.RemoveComponents(m.Name)
	}
}

func (s *App) initRegistry() error {
	if s.Registry == nil {
		return errors.New("registry is not available")
	}
	if s.Scanner != nil {
		s.Registry.SetScanner(s.Scanner)
	}
	s.Registry.Scan()
	return nil
}

func (s *App) wire() error {
	components := s.GetComponents()
	sort.Slice(components, func(i, j int) bool {
		if len(components[i].DependsBy) != len(components[j].DependsBy) {
			return len(components[i].DependsBy) < len(components[j].DependsBy)
		}
		return len(components[i].AllDependencies()) < len(components[j].AllDependencies())
	})
	for _, m := range components {
		err := s.Initialize(s.Registry, m)
		if err != nil {
			return fmt.Errorf("initialize failed: %v", err)
		}
	}
	return nil
}

func (s *App) defaultPostInitFunc(m *meta.Meta) error {
	// before process
	for _, processor := range s.postProcessors {
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
	for _, processor := range s.postProcessors {
		err := processor.PostProcessAfterInitialization(m.Raw)
		if err != nil {
			return fmt.Errorf("post processor: %T process after %s init error: %v", processor, m.ID(), err)
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
