package app

import (
	"context"
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
	"time"
)

type App struct {
	configure.Configure
	container.Factory
	registry           container.SingletonRegistry
	shutdownTimeout    time.Duration
	ApplicationRunners []definition.ApplicationRunner       `wire:",required=false"`
	CloserComponents   []definition.CloserComponent         `wire:",required=false"`
	EventListeners     []definition.ApplicationEventListener `wire:",required=false"`
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

type contextSetter interface {
	SetContext(ctx context.Context)
}

func (s *App) Run(ops ...SettingOption) error {
	return s.RunWithContext(context.Background(), ops...)
}

func (s *App) RunWithContext(ctx context.Context, ops ...SettingOption) error {
	allOps := append(ops, globalOptions...)
	globalOptions = nil
	for _, op := range allOps {
		op(s)
	}
	err := s.initiate()
	if err != nil {
		s.logger().Fatalf("%+v", err)
	}
	if cs, ok := s.Factory.(contextSetter); ok {
		cs.SetContext(ctx)
	}
	if err := s.run(ctx); err != nil {
		s.logger().Errorf("application run failed: %+v", err)
		return err
	}
	return nil
}

func (s *App) run(ctx context.Context) error {
	s.logger().Info("start initializing configuration...")
	if err := s.initConfiguration(); err != nil {
		return errors.WithMessage(err, "application configuration initialize failed")
	}

	s.logger().Info("start initializing component factory...")
	if err := s.initFactory(); err != nil {
		return errors.WithMessage(err, "application factory initialize failed")
	}

	s.logger().Info("start refreshing components...")
	if err := s.refresh(); err != nil {
		return errors.WithMessage(err, "application components refresh failed")
	}

	s.logger().Info("start call up runners...")
	if err := s.callRunners(ctx); err != nil {
		return errors.WithMessagef(err, "start application runners failed")
	}

	_ = s.PublishEvent(&definition.ApplicationStartedEvent{App: s})
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

func (s *App) callRunners(ctx context.Context) error {
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
		var err error
		if r, ok := runner.(definition.ApplicationRunnerWithContext); ok {
			err = r.RunWithContext(ctx)
		} else {
			err = runner.Run()
		}
		if err != nil {
			return errors.Wrapf(err, "invoking Run() for runner '%T'", runner)
		}
	}
	s.ApplicationRunners = nil
	s.logger().Info("all runners started")
	return nil
}

func (s *App) Close() {
	s.CloseWithContext(context.Background())
}

func (s *App) PublishEvent(event definition.ApplicationEvent) error {
	for _, listener := range s.EventListeners {
		if err := listener.OnEvent(event); err != nil {
			s.logger().Errorf("event listener %T failed: %+v", listener, err)
			return errors.Wrapf(err, "event listener %T failed", listener)
		}
	}
	return nil
}

func (s *App) CloseWithContext(ctx context.Context) {
	_ = s.PublishEvent(&definition.ApplicationClosingEvent{App: s})
	if s.shutdownTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.shutdownTimeout)
		defer cancel()
	}
	s.logger().Infof("close closer components")
	if len(s.CloserComponents) != 0 {
		wg := sync.WaitGroup{}
		wg.Add(len(s.CloserComponents))
		for _, m := range s.CloserComponents {
			go func(m definition.CloserComponent) {
				defer wg.Done()
				var err error
				if c, ok := m.(definition.CloserComponentWithContext); ok {
					err = c.CloseWithContext(ctx)
				} else {
					err = m.Close()
				}
				if err != nil {
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
