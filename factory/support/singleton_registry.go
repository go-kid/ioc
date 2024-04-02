package support

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/sync2"
	"github.com/pkg/errors"
)

/*
Registry
Dependency Register and Dependency Lookup
*/

type registry struct {
	componentsMap *sync2.Map[string, any]
}

func (r *registry) GetSingleton(name string) (any, error) {
	c, loaded := r.componentsMap.Load(name)
	if loaded {
		return c, nil
	}
	return nil, errors.Errorf("singleton '%s' not exist", name)
}

func (r *registry) ContainsSingleton(name string) bool {
	_, contains := r.componentsMap.Load(name)
	return contains
}

func (r *registry) GetSingletonNames() []string {
	var names []string
	r.componentsMap.Range(func(key string, _ any) (shouldContinue bool) {
		names = append(names, key)
		return true
	})
	return names
}

func (r *registry) GetSingletonCount() int {
	return len(r.GetSingletonNames())
}

func NewRegistry() SingletonRegistry {
	return &registry{
		componentsMap: sync2.New[string, any](),
	}
}

func (r *registry) RegisterSingleton(singleton any) {
	name := component_definition.GetComponentName(singleton)
	if exist, loaded := r.componentsMap.Load(name); loaded {
		if exist != singleton {
			r.logger().Panicf("register duplicated component %s", name)
		}
		return
	}
	r.componentsMap.Store(name, singleton)
	r.logger().Tracef("singleton registry register component %s", name)
}

func (r *registry) logger() syslog.Logger {
	return syslog.GetLogger().Pref("SingletonRegistry")
}
