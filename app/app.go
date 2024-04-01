package app

import (
	"errors"
	"fmt"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors/definition_registry_post_processors"
	"github.com/go-kid/ioc/factory/processors/instantiation_aware_component_post_processors"
	"github.com/go-kid/ioc/factory/support"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/sort2"
	"sync"
)

type App struct {
	configure.Configure
	factory.Factory
	registry                support.SingletonRegistry
	enableApplicationRunner bool
	ApplicationRunners      []definition.ApplicationRunner `wire:",required=false"`
	CloserComponents        []definition.CloserComponent   `wire:",required=false"`
}

func NewApp(ops ...SettingOption) *App {
	var s = &App{
		Configure:               configure.Default(),
		registry:                support.NewRegistry(),
		Factory:                 factory.Default(),
		enableApplicationRunner: true,
	}
	for _, op := range append(ops, globalOptions...) {
		op(s)
	}
	err := s.initiate()
	if err != nil {
		s.logger().Fatal(err)
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
	var initiateComponent = []any{
		s,
		definition_registry_post_processors.NewPropTagScanProcessor(),
		definition_registry_post_processors.NewValueTagScanProcessor(),
		definition_registry_post_processors.NewWireTagScanProcessor(),
		definition_registry_post_processors.NewFuncTagScanProcessor(),
		instantiation_aware_component_post_processors.NewExpressionTagAwarePostProcessors(),
		instantiation_aware_component_post_processors.NewPropertiesAwarePostProcessors(),
		instantiation_aware_component_post_processors.NewValueAwarePostProcessors(),
		instantiation_aware_component_post_processors.NewValidateAwarePostProcessors(),
		instantiation_aware_component_post_processors.NewRequiredArgValidatePostProcessors(),
		instantiation_aware_component_post_processors.NewDependencyNameAwarePostProcessors(),
		instantiation_aware_component_post_processors.NewDependencyTypeAwarePostProcessors(),
		instantiation_aware_component_post_processors.NewDependencyValidatePostProcessors(),
		instantiation_aware_component_post_processors.NewDependencyFunctionAwarePostProcessors(),
	}
	for _, c := range initiateComponent {
		s.registry.RegisterSingleton(c)
	}
	return nil
}

func (s *App) Run() error {
	if err := s.run(); err != nil {
		s.logger().Errorf("framework run failed: %v", err)
		return err
	}
	return nil
}

func (s *App) run() error {
	/* begin load and bind configuration */
	s.logger().Info("start initializing configuration...")
	if err := s.initConfiguration(); err != nil {
		return fmt.Errorf("init config failed: %v", err)
	}
	s.logger().Info("configuration is loaded")

	/* set default init behavior */
	err := s.initFactory()
	if err != nil {
		return fmt.Errorf("init component factory failed: %v", err)
	}
	s.logger().Info("component factory is ready")
	/* factory ready */

	/* begin inject dependencies */
	s.logger().Info("start refreshing components...")
	if err := s.refresh(); err != nil {
		return fmt.Errorf("refresh components failed: %v", err)
	}
	s.logger().Info("all components is refreshed")
	/* dependency injection ready */

	s.logger().Info("application is issued")

	/* begin call runners */
	if s.enableApplicationRunner {
		s.logger().Info("start starting runners...")
		if err := s.callRunners(); err != nil {
			return fmt.Errorf("start runners failed: %v", err)
		}
	}

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

func (s *App) refresh() error {
	err := s.Factory.Refresh()
	if err != nil {
		return fmt.Errorf("initialize component failed: %v", err)
	}
	return nil
}

func (s *App) callRunners() error {
	runners := s.ApplicationRunners
	if len(runners) == 0 {
		s.logger().Trace("find 0 application runner, skip")
		return nil
	}
	s.logger().Tracef("find %d application runner(s), start sort", len(runners))
	sort2.Slice(runners, func(i, j definition.ApplicationRunner) bool {
		return i.Order() < j.Order()
	})
	for i := range runners {
		runner := runners[i]
		s.logger().Tracef("start runner %T [%d/%d]", runner, i+1, len(runners))
		err := runner.Run()
		if err != nil {
			return fmt.Errorf("start runner %T failed: %v", runner, err)
		}
	}
	s.ApplicationRunners = nil
	s.logger().Info("all runners started")
	return nil
}

func (s *App) Close() {
	if len(s.CloserComponents) != 0 {
		wg := sync.WaitGroup{}
		wg.Add(len(s.CloserComponents))
		for _, m := range s.CloserComponents {
			go func(m definition.CloserComponent) {
				defer wg.Done()
				if err := m.Close(); err != nil {
					s.logger().Errorf("Error closing %T", m)
				}
			}(m)
		}
		wg.Wait()
		s.logger().Infof("close all closer components")
	}
}

func (s *App) logger() syslog.Logger {
	return syslog.GetLogger().Pref("Application")
}
