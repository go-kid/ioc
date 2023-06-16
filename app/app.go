package app

import (
	"fmt"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/meta"
	"github.com/go-kid/ioc/registry"
	"log"
	"sort"
)

type App interface {
	configure.ConfigLoader
	configure.ConfigBinder
	registry.Registry
	factory.Factory
	Run() error
}

type app struct {
	configure.ConfigLoader
	configure.ConfigBinder
	registry.Registry
	factory.Factory
	configPath      string
	postProcessors  []defination.ComponentPostProcessor
	callRunnersFunc func(runners []defination.ApplicationRunner) error
}

func NewApp(ops ...SettingOption) App {
	var s = &app{}
	for _, op := range ops {
		op(s)
	}
	return s
}

func (s *app) Run() error {
	//check registry
	if s.Registry == nil {
		return fmt.Errorf("no registry")
	}
	//init configure
	if s.ConfigLoader == nil {
		s.ConfigLoader = &configure.NopLoader{}
	}
	if s.ConfigBinder == nil {
		s.ConfigBinder = &configure.NopBinder{}
	}
	if s.configPath != "" {
		c, err := s.ConfigLoader.LoadConfig(s.configPath)
		if err != nil {
			return fmt.Errorf("load config failed: %v", err)
		}
		err = s.ConfigBinder.SetConfig(c)
		if err != nil {
			return fmt.Errorf("init config failed: %v", err)
		}
	}
	//init factory
	if s.Factory == nil {
		s.Factory = &factory.DefaultFactory{}
	}
	if s.Factory != nil {
		s.Factory.SetIfNilPreFunc(s.ConfigBinder.PropInject)
		s.Factory.SetIfNilPostInitFunc(s.defaultPostInitFunc)
	}

	s.initComponentPostProcessors()

	for _, m := range s.Registry.GetComponents() {
		err := s.Factory.Initialize(s.Registry, m)
		if err != nil {
			return fmt.Errorf("initialize failed: %v", err)
		}
	}

	err := s.callRunners()
	if err != nil {
		return fmt.Errorf("runners failed: %v", err)
	}
	return nil
}

func (s *app) initComponentPostProcessors() {
	metas := s.Registry.GetBeansByInterface(new(defination.ComponentPostProcessor))
	s.postProcessors = make([]defination.ComponentPostProcessor, 0, len(metas))
	for _, m := range metas {
		err := s.ConfigBinder.PropInject(m)
		if err != nil {
			panic(fmt.Errorf("init post processor %s error: %v", m.ID(), err))
		}
		s.postProcessors = append(s.postProcessors, m.Raw.(defination.ComponentPostProcessor))
		s.Registry.RemoveComponents(m.Name)
	}
}

func (s *app) defaultPostInitFunc(m *meta.Meta) error {
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

func (s *app) callRunners() error {
	metas := s.Registry.GetBeansByInterface(new(defination.ApplicationRunner))
	var runners []defination.ApplicationRunner
	for i := range metas {
		runners = append(runners, metas[i].Raw.(defination.ApplicationRunner))
	}
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
