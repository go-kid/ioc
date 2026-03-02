package support

import (
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/framework_helper"
	"github.com/go-kid/ioc/util/sync2"
	"github.com/pkg/errors"
	"reflect"
)

/*
Registry
Dependency Register and Dependency Lookup
*/

type registry struct {
	componentsMap  *sync2.Map[string, any]
	constructorMap *sync2.Map[string, any]
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

func NewRegistry() container.SingletonRegistry {
	return &registry{
		componentsMap:  sync2.New[string, any](),
		constructorMap: sync2.New[string, any](),
	}
}

func (r *registry) RegisterSingleton(singleton any) {
	t := reflect.TypeOf(singleton)
	if t.Kind() == reflect.Func {
		singleton = r.registerConstructor(singleton, t)
	}
	name := framework_helper.GetComponentName(singleton)
	if exist, loaded := r.componentsMap.Load(name); loaded {
		if exist != singleton {
			r.logger().Panicf("register duplicated component %s", name)
		}
		return
	}
	r.componentsMap.Store(name, singleton)
	r.logger().Tracef("singleton registry register component %s", name)
}

func (r *registry) registerConstructor(constructor any, t reflect.Type) any {
	returnType, err := validateConstructor(t)
	if err != nil {
		r.logger().Panicf("register constructor %s failed: %v", t.String(), err)
	}
	zeroInstance := reflect.New(returnType.Elem()).Interface()
	name := framework_helper.GetComponentName(zeroInstance)
	r.constructorMap.Store(name, constructor)
	r.logger().Tracef("singleton registry register constructor %s for component %s", t.String(), name)
	return zeroInstance
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func validateConstructor(t reflect.Type) (reflect.Type, error) {
	switch t.NumOut() {
	case 1:
		if t.Out(0).Kind() != reflect.Ptr {
			return nil, errors.Errorf("constructor must return a pointer, got %s", t.Out(0))
		}
		return t.Out(0), nil
	case 2:
		if t.Out(0).Kind() != reflect.Ptr {
			return nil, errors.Errorf("constructor first return value must be a pointer, got %s", t.Out(0))
		}
		if !t.Out(1).Implements(errorType) {
			return nil, errors.Errorf("constructor second return value must be error, got %s", t.Out(1))
		}
		return t.Out(0), nil
	default:
		return nil, errors.Errorf("constructor must return 1 or 2 values, got %d", t.NumOut())
	}
}

func (r *registry) GetConstructor(name string) (any, bool) {
	return r.constructorMap.Load(name)
}

func (r *registry) logger() syslog.Logger {
	return syslog.Pref("SingletonRegistry")
}
