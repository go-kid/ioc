package app

import (
	"errors"
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"sort"
	"sync"
)

type App struct {
	configure.Configure
	factory.Factory
	registry                factory.SingletonRegistry
	enableComponentInit     bool
	enableApplicationRunner bool
	ApplicationRunners      []definition.ApplicationRunner `wire:",required=false"`
	CloserComponents        []definition.CloserComponent   `wire:",required=false"`
}

func NewApp(ops ...SettingOption) *App {
	var s = &App{
		Configure:               configure.Default(),
		registry:                factory.GlobalRegistry(),
		Factory:                 factory.Default(),
		enableComponentInit:     true,
		enableApplicationRunner: true,
	}
	for _, op := range append(ops, globalOptions...) {
		op(s)
	}
	err := s.initiate()
	if err != nil {
		syslog.Fatal(err)
	}
	return s
}

func (s *App) initiate() error {
	if s.Configure == nil {
		return errors.New("missing configure")
	}
	if s.registry == nil {
		return errors.New("missing registry")
	}
	if s.Factory == nil {
		return errors.New("missing factory")
	}
	s.Factory.SetRegistry(s.registry)
	s.Factory.SetConfigure(s.Configure)
	s.registry.RegisterSingleton("ApplicationContext", s)
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
	//syslog.Info("start scanning registered component...")
	//syslog.Info("component scanning finished, registry ready")
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
	return nil
}

func (s *App) initFactory() error {
	err := s.Factory.PrepareComponents()
	if err != nil {
		return fmt.Errorf("prepare components error: %v", err)
	}
	return nil
}

func (s *App) wire() error {
	err := s.Factory.Refresh()
	if err != nil {
		return fmt.Errorf("initialize component failed: %v", err)
	}
	return nil
}

func (s *App) callRunners() error {
	runners := s.ApplicationRunners
	if len(runners) == 0 {
		syslog.Trace("find 0 application runner(s), skip")
		return nil
	}
	syslog.Tracef("find %d application runner(s), start sort", len(runners))
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
	s.ApplicationRunners = nil
	return nil
}

func (s *App) Close() {
	wg := sync.WaitGroup{}
	wg.Add(len(s.CloserComponents))
	for _, m := range s.CloserComponents {
		go func(m definition.CloserComponent) {
			defer wg.Done()
			if err := m.Close(); err != nil {
				syslog.Errorf("Error closing %s", component_definition.ComponentId(m))
			}
		}(m)
	}
	wg.Wait()
	syslog.Infof("close all closer components")
}
