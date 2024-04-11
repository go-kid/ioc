package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/framework_helper"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/pkg/errors"
)

type PostProcessorRegistrationDelegate struct {
	componentPostProcessors                     []processors.ComponentPostProcessor
	hasInstantiationAwareComponentPostProcessor bool
	hasDestructionAwareComponentPostProcessor   bool
}

func (f *PostProcessorRegistrationDelegate) RegisterComponentPostProcessors(ps processors.ComponentPostProcessor) {
	switch ps.(type) {
	case processors.InstantiationAwareComponentPostProcessor:
		f.hasInstantiationAwareComponentPostProcessor = true
	case processors.DestructionAwareComponentPostProcessor:
		f.hasDestructionAwareComponentPostProcessor = true
	}
	f.componentPostProcessors = append(f.componentPostProcessors, ps)
}

func (f *PostProcessorRegistrationDelegate) InvokeBeanFactoryPostProcessors(factory Factory, factoryProcessors []ComponentFactoryPostProcessor) error {
	for _, processor := range factoryProcessors {
		err := processor.PostProcessComponentFactory(factory)
		if err != nil {
			return errors.Wrapf(err, "apply %T.PostProcessComponentFactory() for factory %T", processor, factory)
		}
	}
	f.componentPostProcessors = framework_helper.SortOrderedComponents(f.componentPostProcessors)
	for name, component := range factory.GetRegisteredComponents() {
		err := f.applyDefinitionRegistryPostProcessors(factory, component, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *PostProcessorRegistrationDelegate) applyDefinitionRegistryPostProcessors(factory Factory, component any, name string) error {
	for _, processor := range factory.GetDefinitionRegistryPostProcessors() {
		err := processor.PostProcessDefinitionRegistry(factory.GetDefinitionRegistry(), component, name)
		if err != nil {
			return errors.Wrapf(err, "apply %T.PostProcessDefinitionRegistry() for component '%s'", processor, name)
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
		if ipb, ok := processor.(processors.InstantiationAwareComponentPostProcessor); ok {
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
		if ipb, ok := processor.(processors.InstantiationAwareComponentPostProcessor); ok {
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
			if ibp, ok := processor.(processors.SmartInstantiationAwareBeanPostProcessor); ok {
				exposedComponent, err = ibp.GetEarlyBeanReference(exposedComponent, name)
				if err != nil {
					return nil, errors.Wrapf(err, "apply %T.GetEarlyBeanReference() for component '%s'", ibp, name)
				}
			}
		}
	}
	return exposedComponent, nil
}
