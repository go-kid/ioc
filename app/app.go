package app

import (
	"fmt"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/meta"
	"github.com/go-kid/ioc/registry"
	"github.com/samber/lo"
	"log"
	"sort"
)

type App struct {
	configure.ConfigLoader
	configure.ConfigBinder
	registry.Registry
	factory.Factory
	configPath      string
	postProcessors  []defination.ComponentPostProcessor
	callRunnersFunc func(runners []defination.ApplicationRunner) error
}

func NewApp(ops ...SettingOption) *App {
	var s = &App{
		ConfigLoader:    &configure.NopLoader{},
		ConfigBinder:    &configure.NopBinder{},
		Registry:        nil,
		Factory:         &factory.DefaultFactory{},
		configPath:      "",
		postProcessors:  nil,
		callRunnersFunc: nil,
	}
	for _, op := range ops {
		op(s)
	}
	return s
}

func (s *App) Run() error {
	//check registry
	if s.Registry == nil {
		return fmt.Errorf("no registry")
	}
	s.initProduceComponents()
	//init configure
	if err := s.initConfig(); err != nil {
		return err
	}
	//init factory
	s.SetIfNilPostInitFunc(s.defaultPostInitFunc)

	s.initComponentPostProcessors()

	if err := s.wire(); err != nil {
		return err
	}

	if err := s.callRunners(); err != nil {
		return fmt.Errorf("runners failed: %v", err)
	}
	return nil
}

func (s *App) initProduceComponents() {
	metas := s.GetComponents()
	produces := lo.FlatMap[*meta.Meta, *meta.Meta](metas, func(item *meta.Meta, _ int) []*meta.Meta {
		return item.Produce
	})
	lo.ForEach(produces, func(item *meta.Meta, _ int) {
		s.Register(item)
	})
}

func (s *App) initConfig() error {
	if s.configPath == "" {
		return nil
	}
	c, err := s.LoadConfig(s.configPath)
	if err != nil {
		return fmt.Errorf("load config failed: %v", err)
	}
	err = s.SetConfig(c)
	if err != nil {
		return fmt.Errorf("init config failed: %v", err)
	}
	metas := s.GetComponents()
	for _, m := range metas {
		err = s.PropInject(m.Properties)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *App) initComponentPostProcessors() {
	metas := s.GetBeansByInterface(new(defination.ComponentPostProcessor))
	s.postProcessors = make([]defination.ComponentPostProcessor, 0, len(metas))
	for _, m := range metas {
		s.postProcessors = append(s.postProcessors, m.Raw.(defination.ComponentPostProcessor))
		s.RemoveComponents(m.Name)
	}
}

func (s *App) wire() error {
	components := s.GetComponents()
	sort.Slice(components, func(i, j int) bool {
		if len(components[i].DependsBy) != len(components[j].DependsBy) {
			return len(components[i].DependsBy) < len(components[j].DependsBy)
		}
		return len(components[i].Dependencies) < len(components[j].Dependencies)
	})
	for _, m := range components {
		err := s.Initialize(s, m)
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
	log.Printf("ioc: %s inited\n", m.ID())

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
	metas := s.GetBeansByInterface(new(defination.ApplicationRunner))
	var runners = lo.Map(metas, func(item *meta.Meta, _ int) defination.ApplicationRunner {
		return item.Raw.(defination.ApplicationRunner)
	})
	sort.Slice(runners, func(i, j int) bool {
		return runners[i].Order() < runners[j].Order()
	})
	if s.callRunnersFunc == nil {
		s.callRunnersFunc = func(runners []defination.ApplicationRunner) error {
			for i := range runners {
				err := runners[i].Run()
				if err != nil {
					return err
				}
			}
			return nil
		}
	}
	return s.callRunnersFunc(runners)
}
