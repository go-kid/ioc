package factory

import (
	"fmt"
	"github.com/kidhat/kid-ioc/configure"
	"github.com/kidhat/kid-ioc/defination"
	"github.com/kidhat/kid-ioc/injector"
	"github.com/kidhat/kid-ioc/meta"
	"github.com/kidhat/kid-ioc/registry"
	"reflect"
	"sort"
)

type Factory interface {
	Start() error
	Wire(m *meta.Meta) error
}

type factory struct {
	r              *registry.Registry
	c              *configure.Configure
	postProcessors []defination.ComponentPostProcessor
}

func NewFactory(r *registry.Registry, c *configure.Configure) Factory {
	return &factory{
		r: r,
		c: c,
	}
}

func (f *factory) Start() error {
	f.initComponentPostProcessors()
	for _, m := range f.r.GetComponents() {
		err := f.Wire(m)
		if err != nil {
			return fmt.Errorf("initialize failed: %v", err)
		}
	}
	err := f.callRunners()
	if err != nil {
		return fmt.Errorf("runners failed: %v", err)
	}
	return nil
}

func (f *factory) Wire(m *meta.Meta) error {
	if f.r.IsComponentInited(m.Name) {
		return nil
	}

	err := f.c.PropInject(m)
	if err != nil {
		return err
	}

	err = injector.DependencyInject(f.r, m)
	if err != nil {
		return err
	}
	f.r.ComponentInited(m.Name)

	for _, dependency := range m.Dependencies {
		switch dependency.Type.Kind() {
		case reflect.Slice:
			for i := 0; i < dependency.Value.Len(); i++ {
				elem := dependency.Value.Index(i)
				name := defination.GetComponentName(elem.Interface())
				dm := f.r.GetComponentByName(name)
				dm.DependBy(m)
				err := f.Wire(dm)
				if err != nil {
					return err
				}
			}
		default:
			dm := f.r.GetComponentByName(dependency.Name())
			dm.DependBy(m)
			err := f.Wire(dm)
			if err != nil {
				return err
			}
		}
	}

	err = f.applyPostProcessors(m)
	if err != nil {
		return err
	}
	return nil
}

func (f *factory) callRunners() error {
	metas := f.r.GetBeansByInterface(new(defination.ApplicationRunner))
	var runners []defination.ApplicationRunner
	for i := range metas {
		runners = append(runners, metas[i].Raw.(defination.ApplicationRunner))
	}
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

func (f *factory) initComponentPostProcessors() {
	metas := f.r.GetBeansByInterface(new(defination.ComponentPostProcessor))
	for _, m := range metas {
		err := f.c.PropInject(m)
		if err != nil {
			panic(fmt.Errorf("init post processor %s error: %v", m.ID(), err))
		}
		f.postProcessors = append(f.postProcessors, m.Raw.(defination.ComponentPostProcessor))
		f.r.RemoveComponents(m.Name)
	}
}

func (f *factory) applyPostProcessors(m *meta.Meta) error {
	// before process
	for _, processor := range f.postProcessors {
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
	//log.Info().Msgf("ioc: %s inited", m.ID())

	// after process
	for _, processor := range f.postProcessors {
		err := processor.PostProcessAfterInitialization(m.Raw)
		if err != nil {
			return fmt.Errorf("post processor: %T process after %s init error: %v", processor, m.ID(), err)
		}
	}

	return nil
}
