package app

import (
	"errors"
	"fmt"
	"github.com/go-kid/ioc/configure"
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
	configure.Configure
	injector.Injector
	registry.Registry
	factory.Factory
	scanner.Scanner
	enableComponentInit     bool
	enableApplicationRunner bool
}

func NewApp(ops ...SettingOption) *App {
	var s = &App{
		Configure:               configure.Default(),
		Injector:                injector.Default(),
		Registry:                registry.GlobalRegistry(),
		Factory:                 factory.Default(),
		Scanner:                 scanner.Default(),
		enableComponentInit:     true,
		enableApplicationRunner: true,
	}
	Options(ops...)(s)
	err := s.validate()
	if err != nil {
		syslog.Fatal(err)
	}
	return s
}

func (s *App) validate() error {
	if s.Configure == nil {
		return errors.New("missing configure")
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
	if err := s.run(); err != nil {
		defer s.Close()
		syslog.Errorf("framework run failed: %v", err)
		return err
	}
	return nil
}

func (s *App) run() error {
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
	syslog.Info("populate finished, components properties ready")

	/* set default init behavior */
	s.Factory.SetIfNilPostInitFunc(s.defaultPostInitFunc())
	syslog.Info("factory ready")
	/* factory ready */

	/* begin inject dependencies */
	syslog.Info("start wire dependencies...")
	if err := s.wire(); err != nil {
		return fmt.Errorf("wire dependencies failed: %v", err)
	}
	syslog.Info("injection finished, components dependencies ready")
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
	metas := s.GetComponents()
	err := s.Configure.Initialize(metas...)
	if err != nil {
		return fmt.Errorf("initialize configure failed: %v", err)
	}
	err = s.Configure.Populate(metas...)
	if err != nil {
		return fmt.Errorf("populate configure failed: %v", err)
	}
	return nil
}

func (s *App) wire() error {
	components := s.Registry.GetComponents()
	err := s.Factory.Initialize(s.Registry, s.Injector, components...)
	if err != nil {
		return fmt.Errorf("initialize component failed: %v", err)
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

func (s *App) defaultPostInitFunc() factory.MetaFunc {
	if !s.enableComponentInit {
		return func(m *meta.Meta) error {
			return nil
		}
	}
	postMetas := s.Registry.GetComponents(registry.Interface(new(defination.ComponentPostProcessor)))
	if len(postMetas) == 0 {
		return func(m *meta.Meta) error {
			// init
			if ic, ok := m.Raw.(defination.InitializeComponent); ok {
				syslog.Tracef("component %s is InitializeComponent, start do init", m.ID())
				err := ic.Init()
				if err != nil {
					return fmt.Errorf("component %s inited failed: %s", m.ID(), err)
				}
			}
			return nil
		}
	}

	postProcessors := make([]defination.ComponentPostProcessor, len(postMetas))
	for i, pm := range postMetas {
		syslog.Tracef("collecting post processors %s", pm.ID())
		postProcessors[i] = pm.Raw.(defination.ComponentPostProcessor)
		s.Registry.RemoveComponents(pm.Name)
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
			syslog.Tracef("component %s is InitializeComponent, start do init", m.ID())
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
