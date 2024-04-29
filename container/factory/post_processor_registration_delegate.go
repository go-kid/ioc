package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/framework_helper"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/pkg/errors"
	"sync"
)

type PostProcessorRegistrationDelegate struct {
	rawComponentPostProcessors                  []container.ComponentPostProcessor
	componentPostProcessors                     []container.ComponentPostProcessor
	hasInstantiationAwareComponentPostProcessor bool
	hasDestructionAwareComponentPostProcessor   bool
}

func NewPostProcessorRegistrationDelegate() *PostProcessorRegistrationDelegate {
	return &PostProcessorRegistrationDelegate{}
}

func (f *PostProcessorRegistrationDelegate) RegisterComponentPostProcessors(ps container.ComponentPostProcessor, name string) {
	switch ps.(type) {
	case container.InstantiationAwareComponentPostProcessor:
		f.hasInstantiationAwareComponentPostProcessor = true
	case container.DestructionAwareComponentPostProcessor:
		f.hasDestructionAwareComponentPostProcessor = true
	}

	f.rawComponentPostProcessors = append(f.rawComponentPostProcessors, ps)
}

func (f *PostProcessorRegistrationDelegate) InvokeBeanFactoryPostProcessors(factory container.Factory, factoryProcessors []container.ComponentFactoryPostProcessor) error {
	for _, processor := range factoryProcessors {
		err := processor.PostProcessComponentFactory(factory)
		if err != nil {
			return errors.Wrapf(err, "apply %T.PostProcessComponentFactory() for factory %T", processor, factory)
		}
	}

	err := f.applyDefinitionRegistryPostProcessors(factory)
	if err != nil {
		return err
	}

	f.rawComponentPostProcessors = framework_helper.SortOrderedComponents(f.rawComponentPostProcessors)
	for _, processor := range f.rawComponentPostProcessors {
		if _, lazy := processor.(definition.LazyInit); !lazy {
			instance, err := factory.GetComponentByName(framework_helper.GetComponentName(processor))
			if err != nil {
				return err
			}
			if icp, ok := instance.(container.ComponentPostProcessor); ok {
				processor = icp
			}
		}
		f.componentPostProcessors = append(f.componentPostProcessors, processor)
	}
	f.rawComponentPostProcessors = nil
	return nil
}

func (f *PostProcessorRegistrationDelegate) applyDefinitionRegistryPostProcessors(factory container.Factory) error {
	components := factory.GetRegisteredComponents()
	var wg sync.WaitGroup
	for _, processor := range factory.GetDefinitionRegistryPostProcessors() {
		var errs []error
		wg.Add(len(components))
		for name, component := range components {
			go func(name string, component any) {
				defer wg.Done()
				err := processor.PostProcessDefinitionRegistry(factory.GetDefinitionRegistry(), component, name)
				if err != nil {
					errs = append(errs, errors.WithMessage(err, name))
				}
			}(name, component)
		}
		wg.Wait()
		if errs != nil {
			var err = errors.Errorf("apply %T.PostProcessDefinitionRegistry() for component", processor)
			for _, e2 := range errs {
				err = errors.WithMessage(err, e2.Error())
			}
			return err
		}
	}
	return nil
}

func (f *PostProcessorRegistrationDelegate) InitializeComponent(name string, m any) (any, error) {
	f.logger().Tracef("start initialize component '%s'", name)
	var wrappedComponent = m
	var err error
	f.logger().Tracef("start to apply post process before component '%s' initialization", name)
	wrappedComponent, err = f.applyPostProcessBeforeInitialization(wrappedComponent, name)
	if err != nil {
		return nil, err
	}
	if wrappedComponent == nil {
		return m, nil
	}
	err = f.invokeInitMethods(name, wrappedComponent)
	if err != nil {
		return nil, err
	}
	f.logger().Tracef("start to apply post process after component '%s' initialization", name)
	wrappedComponent, err = f.applyPostProcessAfterInitialization(wrappedComponent, name)
	if err != nil {
		return nil, err
	}
	return wrappedComponent, nil
}

func (f *PostProcessorRegistrationDelegate) invokeInitMethods(name string, component any) error {
	if ic, ok := component.(definition.InitializingComponent); ok {
		f.logger().Tracef("invoking afterPropertiesSet() method for component '%s'", name)
		err := ic.AfterPropertiesSet()
		if err != nil {
			return errors.Wrapf(err, "invoking AfterPropertiesSet() method for component '%s'", name)
		}
	}
	if c, ok := component.(definition.InitializeComponent); ok {
		f.logger().Tracef("invoking init method for component '%s'", name)
		err := c.Init()
		if err != nil {
			return errors.Wrapf(err, "invoking Init() for '%s'", name)
		}
	}
	return nil
}

func (f *PostProcessorRegistrationDelegate) applyPostProcessBeforeInitialization(c any, name string) (any, error) {
	var (
		current = c
		err     error
	)
	for _, processor := range f.componentPostProcessors {
		current, err = processor.PostProcessBeforeInitialization(current, name)
		if err != nil {
			return nil, errors.Wrapf(err, "component post processor %s apply post process before initialization", reflectx.Id(processor))
		}
		if current == nil {
			return nil, nil
		}
	}
	return current, nil
}

func (f *PostProcessorRegistrationDelegate) applyPostProcessAfterInitialization(c any, name string) (any, error) {
	var (
		result = c
		err    error
	)
	var current any
	for _, processor := range f.componentPostProcessors {
		current, err = processor.PostProcessAfterInitialization(result, name)
		if err != nil {
			return nil, errors.Wrapf(err, "component post processor %s apply post process after initialization", reflectx.Id(processor))
		}
		if current == nil {
			return result, nil
		}
		result = current
	}
	return result, nil
}

func (f *PostProcessorRegistrationDelegate) logger() syslog.Logger {
	return syslog.GetLogger().Pref("PostProcessorDelegate")
}

func (f *PostProcessorRegistrationDelegate) ResolveBeforeInstantiation(meta *component_definition.Meta, name string) (any, error) {
	var component any
	var err error
	if f.hasInstantiationAwareComponentPostProcessor {
		component, err = f.applyPostProcessBeforeInstantiation(meta, name)
		if err != nil {
			return nil, err
		}
		if component != nil {
			component, err = f.applyPostProcessAfterInitialization(component, name)
			if err != nil {
				return nil, err
			}
		}
	}
	return component, nil
}

func (f *PostProcessorRegistrationDelegate) applyPostProcessBeforeInstantiation(meta *component_definition.Meta, name string) (any, error) {
	var component any
	var err error
	for _, processor := range f.componentPostProcessors {
		if ipb, ok := processor.(container.InstantiationAwareComponentPostProcessor); ok {
			component, err = ipb.PostProcessBeforeInstantiation(meta, name)
			if err != nil {
				return nil, errors.Wrapf(err, "apply %T.PostProcessBeforeInstantiation() for component '%s'", ipb, name)
			}
			if component != nil {
				return component, nil
			}
		}
	}
	return component, nil
}

func (f *PostProcessorRegistrationDelegate) ResolveAfterInstantiation(meta *component_definition.Meta, name string) error {
	for _, processor := range f.componentPostProcessors {
		if ipb, ok := processor.(container.InstantiationAwareComponentPostProcessor); ok {
			ok, err := ipb.PostProcessAfterInstantiation(meta.Raw, name)
			if err != nil {
				return errors.Wrapf(err, "apply %T.PostProcessAfterInstantiation() for component '%s'", ipb, name)
			}
			if ok {
				_, err := ipb.PostProcessProperties(meta.GetAllProperties(), meta.Raw, name)
				if err != nil {
					return errors.Wrapf(err, "apply %T.PostProcessProperties() for component '%s'", ipb, name)
				}
				//meta.SetProperties(properties...)
			}
		}
	}
	return nil
}

func (f *PostProcessorRegistrationDelegate) GetEarlyBeanReference(name string, m any) (any, error) {
	var exposedComponent = m
	var err error
	if f.hasInstantiationAwareComponentPostProcessor {
		for _, processor := range f.componentPostProcessors {
			if ibp, ok := processor.(container.SmartInstantiationAwareBeanPostProcessor); ok {
				exposedComponent, err = ibp.GetEarlyBeanReference(exposedComponent, name)
				if err != nil {
					return nil, errors.Wrapf(err, "apply %T.GetEarlyBeanReference() for component '%s'", ibp, name)
				}
			}
		}
	}
	return exposedComponent, nil
}
