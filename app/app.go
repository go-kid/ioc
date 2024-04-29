package app

import (
	"flag"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/container/factory"
	"github.com/go-kid/ioc/container/processors"
	"github.com/go-kid/ioc/container/support"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/framework_helper"
	"github.com/pkg/errors"
	"sync"
)

type App struct {
	configure.Configure
	container.Factory
	registry           container.SingletonRegistry
	ApplicationRunners []definition.ApplicationRunner `wire:",required=false"`
	CloserComponents   []definition.CloserComponent   `wire:",required=false"`
}

func NewApp() *App {
	defer flag.Parse()
	var s = &App{
		Configure: configure.Default(),
		registry:  support.NewRegistry(),
		Factory:   factory.Default(),
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
		processors.NewLoggerAwarePostProcessor(),
		processors.NewConfigQuoteAwarePostProcessors(),
		processors.NewExpressionTagAwarePostProcessors(),
		processors.NewPropertiesAwarePostProcessors(),
		processors.NewValueAwarePostProcessors(),
		processors.NewValidateAwarePostProcessors(),
		processors.NewDependencyAwarePostProcessors(),
		processors.NewDependencyFurtherMatchingProcessors(),
		processors.NewDependencyFunctionAwarePostProcessors(),
	}
	for _, c := range initiateComponent {
		s.registry.RegisterSingleton(c)
	}
	return nil
}

func (s *App) Run(ops ...SettingOption) error {
	for _, op := range append(ops, globalOptions...) {
		op(s)
	}
	err := s.initiate()
	if err != nil {
		s.logger().Fatalf("%+v", err)
	}
	if err := s.run(); err != nil {
		s.logger().Errorf("application run failed: %+v", err)
		return err
	}
	return nil
}

func (s *App) run() error {
	/* begin load and bind configuration */
	s.logger().Info("start initializing configuration...")
	if err := s.initConfiguration(); err != nil {
		return errors.WithMessage(err, "application configuration initialize failed")
	}

	/* set default init behavior */
	s.logger().Info("start initializing component factory...")
	if err := s.initFactory(); err != nil {
		return errors.WithMessage(err, "application factory initialize failed")
	}
	/* factory ready */

	/* begin inject dependencies */
	s.logger().Info("start refreshing components...")
	if err := s.refresh(); err != nil {
		return errors.WithMessage(err, "application components refresh failed")
	}
	/* dependency injection ready */

	/* begin call runners */
	s.logger().Info("start call up runners...")
	if err := s.callRunners(); err != nil {
		return errors.WithMessagef(err, "start application runners failed")
	}
	/* finished */
	s.logger().Info("application run up")
	return nil
}

func (s *App) initConfiguration() error {
	err := s.Configure.Initialize()
	if err != nil {
		return errors.WithMessage(err, "initialize configure error")
	}
	return nil
}

func (s *App) initFactory() error {
	err := s.Factory.PrepareComponents()
	if err != nil {
		return errors.WithMessage(err, "prepare component factory error")
	}
	return nil
}

func (s *App) refresh() error {
	err := s.Factory.Refresh()
	if err != nil {
		return errors.WithMessage(err, "component factory refresh error")
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
	runners = framework_helper.SortOrderedComponents(runners)
	for i := range runners {
		runner := runners[i]
		s.logger().Tracef("start runner %T [%d/%d]", runner, i+1, len(runners))
		err := runner.Run()
		if err != nil {
			return errors.Wrapf(err, "invoking Run() for runner '%T'", runner)
		}
	}
	s.ApplicationRunners = nil
	s.logger().Info("all runners started")
	return nil
}

func (s *App) Close() {
	s.logger().Infof("close closer components")
	if len(s.CloserComponents) != 0 {
		wg := sync.WaitGroup{}
		wg.Add(len(s.CloserComponents))
		for _, m := range s.CloserComponents {
			go func(m definition.CloserComponent) {
				defer wg.Done()
				if err := m.Close(); err != nil {
					err = errors.Wrapf(err, "invoking Close() for closer '%T'", m)
					s.logger().Errorf("%+v", err)
				}
			}(m)
		}
		wg.Wait()
	} else {
		s.logger().Trace("find 0 closer component, skip")
	}
}

func (s *App) logger() syslog.Logger {
	return syslog.Pref("Application")
}
