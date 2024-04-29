package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/container/support"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/sort2"
	"github.com/pkg/errors"
)

type defaultFactory struct {
	singletonRegistry                 container.SingletonRegistry
	definitionRegistry                container.DefinitionRegistry
	singletonComponentRegistry        container.SingletonComponentRegistry
	configure                         configure.Configure
	definitionRegistryPostProcessors  []container.DefinitionRegistryPostProcessor
	allowCircularReferences           bool
	postProcessorRegistrationDelegate *PostProcessorRegistrationDelegate
	registeredComponents              map[string]any
}

func Default() container.Factory {
	f := &defaultFactory{
		definitionRegistry:                support.DefaultDefinitionRegistry(),
		singletonComponentRegistry:        support.DefaultSingletonComponentRegistry(),
		postProcessorRegistrationDelegate: NewPostProcessorRegistrationDelegate(),
		allowCircularReferences:           true,
	}
	return f
}

func (f *defaultFactory) PrepareComponents() error {
	singletonNames := f.singletonRegistry.GetSingletonNames()
	f.registeredComponents = make(map[string]any, len(singletonNames))
	var factoryPostProcessors []container.ComponentFactoryPostProcessor
	for _, name := range singletonNames {
		singleton, err := f.singletonRegistry.GetSingleton(name)
		if err != nil {
			return err
		}
		if p, ok := singleton.(container.ComponentPostProcessor); ok {
			f.registerBeanPostProcessors(p, name)
		}
		if p, ok := singleton.(container.DefinitionRegistryPostProcessor); ok {
			f.definitionRegistryPostProcessors = append(f.definitionRegistryPostProcessors, p)
		}
		if p, ok := singleton.(container.ComponentFactoryPostProcessor); ok {
			factoryPostProcessors = append(factoryPostProcessors, p)
		}
		f.registeredComponents[name] = singleton
	}

	err := f.postProcessorRegistrationDelegate.InvokeBeanFactoryPostProcessors(f, factoryPostProcessors)
	if err != nil {
		return err
	}
	f.logger().Info("prepare components finished")
	return nil
}

func (f *defaultFactory) registerBeanPostProcessors(postProcessor container.ComponentPostProcessor, name string) {
	f.postProcessorRegistrationDelegate.RegisterComponentPostProcessors(postProcessor, name)
}

func (f *defaultFactory) GetRegisteredComponents() map[string]any {
	return f.registeredComponents
}

func (f *defaultFactory) GetDefinitionRegistryPostProcessors() []container.DefinitionRegistryPostProcessor {
	return f.definitionRegistryPostProcessors
}

func (f *defaultFactory) SetRegistry(r container.SingletonRegistry) {
	f.singletonRegistry = r
}

func (f *defaultFactory) SetConfigure(c configure.Configure) {
	f.configure = c
}

func (f *defaultFactory) GetConfigure() configure.Configure {
	return f.configure
}

func (f *defaultFactory) GetDefinitionRegistry() container.DefinitionRegistry {
	return f.definitionRegistry
}

func (f *defaultFactory) Refresh() error {

	var names []string

	for _, meta := range f.definitionRegistry.GetMetas() {
		switch meta.Raw.(type) {
		case definition.LazyInit:
			continue
		default:
			names = append(names, meta.Name())
		}
	}

	sort2.Slice(names, func(i, j string) bool {
		return i < j
	})
	for _, name := range names {
		f.logger().Tracef("refresh component with name '%s'", name)
		_, err := f.doGetComponent(name)
		if err != nil {
			return err
		}
	}

	f.logger().Info("refresh components finished")
	return nil
}

func (f *defaultFactory) GetComponents(opts ...container.Option) ([]any, error) {
	var components []any
	for _, meta := range f.definitionRegistry.GetMetas(opts...) {
		component, err := f.GetComponentByName(meta.Name())
		if err != nil {
			return nil, err
		}
		components = append(components, component)
	}
	return components, nil
}

func (f *defaultFactory) GetComponentByName(name string) (any, error) {
	m, err := f.doGetComponent(name)
	if err != nil {
		return nil, err
	}
	return m.Raw, nil
}

func (f *defaultFactory) doGetComponent(name string) (*component_definition.Meta, error) {
	sharedInstance, err := f.singletonComponentRegistry.GetSingleton(name, true)
	if err != nil {
		return nil, err
	}
	if sharedInstance != nil {
		if f.singletonComponentRegistry.IsSingletonCurrentlyInCreation(name) {
			f.logger().Debugf("returning eagerly cached instance of singleton '%s' that is not fully initialized yet - a consequence of a circular reference",
				name)
		} else {
			f.logger().Debugf("returning eagerly cached instance of singleton '%s'", name)
		}
		return sharedInstance, nil
	}
	sharedInstance, err = f.singletonComponentRegistry.GetSingletonOrCreateByFactory(name,
		container.FuncSingletonFactory(func() (*component_definition.Meta, error) {
			return f.createComponent(name)
		}))
	if err != nil {
		return nil, err
	}
	return sharedInstance, nil
}

func (f *defaultFactory) createComponent(name string) (*component_definition.Meta, error) {
	meta := f.definitionRegistry.GetMetaByName(name)
	if meta == nil {
		return nil, errors.Errorf("component definition with name '%s' not found", name)
	}

	instantiation, err := f.postProcessorRegistrationDelegate.ResolveBeforeInstantiation(meta, name)
	if err != nil {
		return nil, err
	}
	if instantiation != nil {
		if instantiation != meta.Raw {
			return component_definition.CreateProxy(meta, name, instantiation)
		} else {
			return meta, nil
		}
	}

	instance, err := f.doCreateComponent(name, meta)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (f *defaultFactory) doCreateComponent(name string, meta *component_definition.Meta) (*component_definition.Meta, error) {
	// set to singleton component factory cache
	earlySingletonExposure := meta.IsSingleton() && f.allowCircularReferences && f.singletonComponentRegistry.IsSingletonCurrentlyInCreation(name)
	if earlySingletonExposure {
		f.logger().Debugf("eagerly caching bean '%s' to allow for resolving potential circular references", name)
		f.singletonComponentRegistry.AddSingletonFactory(name, container.FuncSingletonFactory(func() (*component_definition.Meta, error) {
			return f.getEarlyBeanReference(name, meta)
		}))
	}

	var exposedComponent = meta

	err := f.populateComponent(name, meta)
	if err != nil {
		return nil, err
	}

	instance := meta.Raw
	wrappedInstance, err := f.postProcessorRegistrationDelegate.InitializeComponent(name, instance)
	if err != nil {
		return nil, err
	}
	if wrappedInstance != instance {
		exposedComponent, err = f.genProxyComponent(meta, name, wrappedInstance)
		if err != nil {
			return nil, err
		}
		f.logger().Debugf("component '%s' initialization finished, detected a new instance '%s' proxy, return proxy component", name, exposedComponent.Type.String())
	} else {
		f.logger().Debugf("component '%s' initialization finished", name)
	}

	if earlySingletonExposure {
		f.logger().Tracef("try get early singleton reference '%s' to check with currently exposed component", name)
		earlySingletonReference, err := f.singletonComponentRegistry.GetSingleton(name, false)
		if err != nil {
			return nil, err
		}
		if earlySingletonReference != nil {
			f.logger().Tracef("early singleton reference for '%s' exists, start check other dependents reference version", name)
			if exposedComponent == meta {
				exposedComponent = earlySingletonReference
				f.logger().Tracef("early singleton reference for '%s' is equal to currently exposed component, use early singleton reference to exposed", name)
			} else if dependents := append(earlySingletonReference.GetDependents(), meta.GetDependents()...); len(dependents) != 0 {
				f.logger().Tracef("early singleton reference with name '%s' has been injected into components %s", name, dependents)
				var actualDependents []string
				for _, dependent := range dependents {
					if !f.singletonComponentRegistry.IsSingletonCurrentlyInCreation(dependent) {
						actualDependents = append(actualDependents, dependent)
					}
				}
				if len(actualDependents) != 0 {
					return nil, errors.Errorf("singleton with name '%s' has been injected into other components \n%s, but has been wrapped which means that other beans do not use the final version of the bean, please try change component init order.",
						name, actualDependents)
				}
			}
		}
	}
	f.logger().Debugf("do create component '%s' finished", name)
	return exposedComponent, nil
}

func (f *defaultFactory) populateComponent(name string, meta *component_definition.Meta) error {
	err := f.postProcessorRegistrationDelegate.ResolveAfterInstantiation(meta, name)
	if err != nil {
		return err
	}
	if properties := meta.GetComponentProperties(); len(properties) > 0 {
		f.logger().Tracef("inject dependencies for '%s'", name)
		for _, node := range meta.GetComponentProperties() {
			if dependencies := node.Injects; len(dependencies) != 0 {
				var injects []*component_definition.Meta
				for _, dependency := range node.Injects {
					f.logger().Tracef("found dependency '%s' for '%s', start to get or create", dependency.Name(), name)
					component, err := f.doGetComponent(dependency.Name())
					if err != nil {
						return err
					}
					injects = append(injects, component)
				}
				err = node.Inject(injects)
				if err != nil {
					return err
				}
			}
		}
		f.logger().Tracef("finished inject dependencies for '%s'", name)
	}
	f.logger().Debugf("component '%s' population finished", name)
	return nil
}

func (f *defaultFactory) getEarlyBeanReference(name string, m *component_definition.Meta) (*component_definition.Meta, error) {
	var exposedComponent = m.Raw
	var err error
	exposedComponent, err = f.postProcessorRegistrationDelegate.GetEarlyBeanReference(name, exposedComponent)
	if err != nil {
		return nil, err
	}
	if exposedComponent != m.Raw {
		m, err = f.genProxyComponent(m, name, exposedComponent)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (f *defaultFactory) genProxyComponent(origin *component_definition.Meta, name string, newComponent any) (*component_definition.Meta, error) {
	return component_definition.CreateProxy(origin, name, newComponent)
}

func (f *defaultFactory) logger() syslog.Logger {
	return syslog.GetLogger().Pref("ComponentFactory")
}
