package app

import (
	"errors"
	"fmt"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/factory"
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
	registry.Registry
	factory.Factory
	scanner.Scanner
	enableComponentInit     bool
	enableApplicationRunner bool
}

func NewApp(ops ...SettingOption) *App {
	var s = &App{
		Configure:               configure.Default(),
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
	if err := s.initConfiguration(); err != nil {
		return fmt.Errorf("init config failed: %v", err)
	}
	syslog.Info("populate finished, components properties ready")

	/* set default init behavior */
	err := s.initFactory()
	if err != nil {
		return fmt.Errorf("init factory failed: %v", err)
	}
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

func (s *App) initConfiguration() error {
	err := s.Configure.Initialize()
	if err != nil {
		return fmt.Errorf("initialize configure failed: %v", err)
	}
	syslog.Trace("start populating properties...")
	err = s.Configure.PopulateProperties(s.getStarters()...)
	if err != nil {
		return fmt.Errorf("populate components properties: %v", err)
	}
	syslog.Info("populate properties finished")
	return nil
}

func (s *App) initFactory() error {
	s.Factory.SetRegistry(s.Registry)
	s.Factory.SetConfigure(s.Configure)
	err := s.Factory.PrepareSpecialComponents()
	if err != nil {
		return fmt.Errorf("prepare special components error: %v", err)
	}
	return nil
}

func (s *App) wire() error {
	components := s.getStarters()
	err := s.Factory.Initialize(components...)
	if err != nil {
		return fmt.Errorf("initialize component failed: %v", err)
	}
	return nil
}

func (s *App) getStarters() []*meta.Meta {
	//return s.Registry.GetComponents(registry.Interface(new(defination.ApplicationStarter)))
	return s.Registry.GetComponents()
}

func (s *App) callRunners() error {
	if !s.enableApplicationRunner {
		return nil
	}
	metas := s.GetComponents(registry.Interface(new(defination.ApplicationRunner)))
	//err := s.Factory.Initialize(metas...)
	//if err != nil {
	//	return fmt.Errorf("initialize application runners failed: %v", err)
	//}
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
