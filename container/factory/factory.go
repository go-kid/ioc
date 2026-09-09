package factory

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/container/support"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/syslog"
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
	ctx                               context.Context
	resolveStack                      []string
	factoryHook                       container.FactoryHook
}

func (f *defaultFactory) SetFactoryHook(hook container.FactoryHook) {
	f.factoryHook = hook
	f.postProcessorRegistrationDelegate.factoryHook = hook
}

func (f *defaultFactory) emitEvent(phase, action, componentName, processorName string, details map[string]any) {
	if f.factoryHook == nil {
		return
	}
	f.factoryHook.OnFactoryEvent(container.FactoryEvent{
		Phase:         phase,
		Action:        action,
		ComponentName: componentName,
		ProcessorName: processorName,
		Details:       details,
	})
}

func (f *defaultFactory) SetContext(ctx context.Context) {
	f.ctx = ctx
}

func (f *defaultFactory) getContext() context.Context {
	if f.ctx != nil {
		return f.ctx
	}
	return context.Background()
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
	f.emitEvent("prepare", "phase_start", "", "", map[string]any{"phase": "PrepareComponents"})

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
		f.emitEvent("prepare", "component_registered", name, "", map[string]any{
			"type": reflect.TypeOf(singleton).String(),
		})
	}

	err := f.postProcessorRegistrationDelegate.InvokeBeanFactoryPostProcessors(f, factoryPostProcessors)
	if err != nil {
		return err
	}
	f.emitEvent("prepare", "phase_end", "", "", map[string]any{"phase": "PrepareComponents"})
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
	f.emitEvent("refresh", "phase_start", "", "", map[string]any{"phase": "Refresh"})

	var names []string
	for _, meta := range f.definitionRegistry.GetMetas() {
		switch meta.Raw.(type) {
		case definition.LazyInit:
			continue
		default:
			if cc, ok := meta.Raw.(definition.ConditionalComponent); ok {
				if !cc.Condition(f.newConditionContext()) {
					f.logger().Debugf("skip conditional component '%s'", meta.Name())
					continue
				}
			}
			names = append(names, meta.Name())
		}
	}

	slices.Sort(names)
	for _, name := range names {
		f.logger().Tracef("refresh component with name '%s'", name)
		f.emitEvent("refresh", "component_resolving", name, "", nil)
		_, err := f.doGetComponent(name)
		if err != nil {
			return err
		}
	}

	f.emitEvent("refresh", "phase_end", "", "", map[string]any{"phase": "Refresh"})
	f.logger().Info("refresh components finished")
	return nil
}

type conditionContext struct {
	registry  container.SingletonRegistry
	configure configure.Configure
}

func (f *defaultFactory) newConditionContext() definition.ConditionContext {
	return &conditionContext{
		registry:  f.singletonRegistry,
		configure: f.configure,
	}
}

func (c *conditionContext) HasComponent(name string) bool {
	return c.registry.ContainsSingleton(name)
}

func (c *conditionContext) GetConfig(key string) interface{} {
	if c.configure == nil {
		return nil
	}
	return c.configure.Get(key)
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

func (f *defaultFactory) pushResolveStack(name string) {
	f.resolveStack = append(f.resolveStack, name)
}

func (f *defaultFactory) popResolveStack() {
	if len(f.resolveStack) > 0 {
		f.resolveStack = f.resolveStack[:len(f.resolveStack)-1]
	}
}

func (f *defaultFactory) formatDependencyChain(failedName string, reason string) string {
	var sb strings.Builder
	sb.WriteString("dependency resolution failed:\n")
	for i, name := range f.resolveStack {
		sb.WriteString(strings.Repeat("  ", i))
		if i > 0 {
			sb.WriteString("-> ")
		}
		sb.WriteString(name)
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Repeat("  ", len(f.resolveStack)))
	sb.WriteString("-> ")
	sb.WriteString(failedName)
	sb.WriteString(fmt.Sprintf(" (%s)", reason))
	return sb.String()
}

func (f *defaultFactory) doGetComponent(name string) (*component_definition.Meta, error) {
	meta := f.definitionRegistry.GetMetaByName(name)
	if meta != nil && meta.IsPrototype() {
		f.logger().Debugf("creating new prototype instance for '%s'", name)
		return f.createComponent(name)
	}

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

	f.pushResolveStack(name)
	sharedInstance, err = f.singletonComponentRegistry.GetSingletonOrCreateByFactory(name,
		container.FuncSingletonFactory(func() (*component_definition.Meta, error) {
			return f.createComponent(name)
		}))
	f.popResolveStack()
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

	f.emitEvent("refresh", "component_creating", name, "", map[string]any{"type": meta.Type.String()})

	f.emitEvent("refresh", "before_instantiation", name, "", nil)
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
	earlySingletonExposure := meta.IsSingleton() && !meta.IsPrototype() && f.allowCircularReferences && f.singletonComponentRegistry.IsSingletonCurrentlyInCreation(name)
	if earlySingletonExposure {
		f.logger().Debugf("eagerly caching bean '%s' to allow for resolving potential circular references", name)
		f.singletonComponentRegistry.AddSingletonFactory(name, container.FuncSingletonFactory(func() (*component_definition.Meta, error) {
			return f.getEarlyBeanReference(name, meta)
		}))
	}

	var exposedComponent = meta

	f.emitEvent("refresh", "populating", name, "", nil)
	err := f.populateComponent(name, meta)
	if err != nil {
		return nil, err
	}
	f.emitEvent("refresh", "populated", name, "", nil)

	instance := meta.Raw
	wrappedInstance, err := f.postProcessorRegistrationDelegate.InitializeComponentWithContext(f.getContext(), name, instance)
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
	f.emitEvent("refresh", "component_ready", name, "", nil)

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
						return fmt.Errorf("%s\n%w", f.formatDependencyChain(dependency.Name(), "not found or creation failed"), err)
					}
					injects = append(injects, component)

					depType := "pointer"
					if node.Type.Kind() == reflect.Interface ||
						(node.Type.Kind() == reflect.Slice && node.Type.Elem().Kind() == reflect.Interface) {
						depType = "interface"
					}
					f.emitEvent("refresh", "dependency_injected", name, "", map[string]any{
						"dependency": dependency.Name(),
						"field":      node.StructField.Name,
						"depType":    depType,
					})
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
	return syslog.Pref("ComponentFactory")
}
