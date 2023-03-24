package ioc

import (
	"fmt"
	"reflect"
	"sort"
)

/*
lifecycle management
*/
func initialize(m *meta, r *registry) error {
	if r.IsComponentInited(m.Name) {
		return nil
	}

	err := propInject(m)
	if err != nil {
		return err
	}

	err = dependencyInject(r, m)
	if err != nil {
		return err
	}
	r.ComponentInited(m.Name)

	for _, dependency := range m.Dependencies {
		switch dependency.Type.Kind() {
		case reflect.Slice:
			for i := 0; i < dependency.Value.Len(); i++ {
				elem := dependency.Value.Index(i)
				name := getComponentName(elem.Interface())
				dm := r.GetComponentByName(name)
				dm.DependBy(m)
				err := initialize(dm, r)
				if err != nil {
					return err
				}
			}
		default:
			dm := r.GetComponentByName(dependency.Name())
			dm.DependBy(m)
			err := initialize(dm, r)
			if err != nil {
				return err
			}
		}
	}

	err = applyPostProcessors(m)
	if err != nil {
		return err
	}
	return nil
}

func callRunners(r *registry) error {
	metas := r.GetBeansByInterface(new(ApplicationRunner))
	var runners []ApplicationRunner
	for i := range metas {
		runners = append(runners, metas[i].Raw.(ApplicationRunner))
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

var postProcessors []ComponentPostProcessor

func initComponentPostProcessors(r *registry) {
	metas := r.GetBeansByInterface(new(ComponentPostProcessor))
	for _, m := range metas {
		err := propInject(m)
		if err != nil {
			panic(fmt.Errorf("init post processor %s error: %v", m.ID(), err))
		}
		postProcessors = append(postProcessors, m.Raw.(ComponentPostProcessor))
		r.RemoveComponents(m.Name)
	}
}

func applyPostProcessors(m *meta) error {
	// before process
	for _, processor := range postProcessors {
		err := processor.PostProcessBeforeInitialization(m.Raw)
		if err != nil {
			return fmt.Errorf("post processor: %T process before %s init error: %v", processor, m.ID(), err)
		}
	}
	// init
	if ic, ok := m.Raw.(InitializeComponent); ok {
		err := ic.Init()
		if err != nil {
			return fmt.Errorf("component: %s inited failed: %s", m.ID(), err)
		}
	}
	//log.Info().Msgf("ioc: %s inited", m.ID())

	// after process
	for _, processor := range postProcessors {
		err := processor.PostProcessAfterInitialization(m.Raw)
		if err != nil {
			return fmt.Errorf("post processor: %T process after %s init error: %v", processor, m.ID(), err)
		}
	}

	return nil
}
