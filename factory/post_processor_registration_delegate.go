package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/ioc/util/sort2"
	"github.com/pkg/errors"
)

type PostProcessorRegistrationDelegate struct {
	priorityOrderedComponentPostProcessors      []processors.ComponentPostProcessor
	orderedComponentPostProcessors              []processors.ComponentPostProcessor
	otherComponentPostProcessors                []processors.ComponentPostProcessor
	componentPostProcessors                     []processors.ComponentPostProcessor
	hasInstantiationAwareComponentPostProcessor bool
	hasDestructionAwareComponentPostProcessor   bool
	hasComponentInitializedPostProcessor        bool
}

func (f *PostProcessorRegistrationDelegate) RegisterComponentPostProcessors(ps processors.ComponentPostProcessor) {
	switch ps.(type) {
	case processors.InstantiationAwareComponentPostProcessor:
		f.hasInstantiationAwareComponentPostProcessor = true
	case processors.DestructionAwareComponentPostProcessor:
		f.hasDestructionAwareComponentPostProcessor = true
	case processors.ComponentInitializedPostProcessor:
		f.hasComponentInitializedPostProcessor = true
	}
	switch ps.(type) {
	case definition.PriorityOrdered:
		f.priorityOrderedComponentPostProcessors = append(f.priorityOrderedComponentPostProcessors, ps)
	case definition.Ordered:
		f.orderedComponentPostProcessors = append(f.orderedComponentPostProcessors, ps)
	default:
		f.otherComponentPostProcessors = append(f.otherComponentPostProcessors, ps)
	}
}

func (f *PostProcessorRegistrationDelegate) InvokeBeanFactoryPostProcessors(factory Factory, factoryProcessors []ComponentFactoryPostProcessor) error {
	for _, processor := range factoryProcessors {
		err := processor.PostProcessComponentFactory(factory)
		if err != nil {
			return err
		}
	}

	sort2.Slice(f.priorityOrderedComponentPostProcessors, func(i processors.ComponentPostProcessor, j processors.ComponentPostProcessor) bool {
		return i.(definition.PriorityOrdered).Order() < j.(definition.PriorityOrdered).Order()
	})
	f.componentPostProcessors = append(f.componentPostProcessors, f.priorityOrderedComponentPostProcessors...)
	sort2.Slice(f.orderedComponentPostProcessors, func(i processors.ComponentPostProcessor, j processors.ComponentPostProcessor) bool {
		return i.(definition.Ordered).Order() < j.(definition.Ordered).Order()
	})
	f.componentPostProcessors = append(f.componentPostProcessors, f.orderedComponentPostProcessors...)
	f.componentPostProcessors = append(f.componentPostProcessors, f.otherComponentPostProcessors...)

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
	return nil
}

func (f *PostProcessorRegistrationDelegate) applyPostProcessBeforeInitialization(c any, name string) (any, error) {
	var (
		result = c
		err    error
	)
	var current any
	for _, processor := range f.componentPostProcessors {
		current, err = processor.PostProcessBeforeInitialization(result, name)
		if err != nil {
			return nil, errors.Wrapf(err, "component post processor %s apply post process before initialization", reflectx.Id(processor))
		}
		if current == nil {
			return result, nil
		}
		result = current
	}
	return result, nil
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

func (f *PostProcessorRegistrationDelegate) ProcessInitializedComponentInitialization(meta *component_definition.Meta) error {
	if f.hasComponentInitializedPostProcessor {
		f.logger().Tracef("post process before initialize component '%s'", meta.Name())
		for _, processor := range f.componentPostProcessors {
			if pb, ok := processor.(processors.ComponentInitializedPostProcessor); ok {
				err := pb.PostProcessBeforeInitialized(meta.Raw)
				if err != nil {
					return errors.Wrapf(err, "apply %T.PostProcessBeforeInitialized() for component '%s'", processor, meta.Name())
				}
			}
		}
	}
	if c, ok := meta.Raw.(definition.InitializeComponent); ok {
		f.logger().Tracef("invoking init method for component '%s'", meta.Name())
		err := c.Init()
		if err != nil {
			return errors.Wrapf(err, "invoking Init() for '%s'", meta.Name())
		}
	}
	if f.hasComponentInitializedPostProcessor {
		f.logger().Tracef("post process after initialize component '%s'", meta.Name())
		for _, processor := range f.componentPostProcessors {
			if pb, ok := processor.(processors.ComponentInitializedPostProcessor); ok {
				err := pb.PostProcessAfterInitialized(meta.Raw)
				if err != nil {
					return errors.Wrapf(err, "apply %T.PostProcessAfterInitialized() for component '%s'", processor, meta.Name())
				}
			}
		}
	}
	return nil
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
