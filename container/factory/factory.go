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
	"github.com/mitchellh/mapstructure"
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
		_, err := f.doGetComponent(name)
		if err != nil {
			return err
		}
	}

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

	if constructor, ok := f.singletonRegistry.GetConstructor(name); ok {
		instance, err := f.invokeConstructor(name, constructor)
		if err != nil {
			return nil, errors.Wrapf(err, "invoke constructor for component '%s'", name)
		}
		meta, err = component_definition.CreateProxy(meta, name, instance)
		if err != nil {
			return nil, err
		}
		f.definitionRegistry.RegisterMeta(meta)
		for _, processor := range f.definitionRegistryPostProcessors {
			if err := processor.PostProcessDefinitionRegistry(f.definitionRegistry, meta.Raw, name); err != nil {
				return nil, errors.Wrapf(err, "re-scan definition for constructor component '%s'", name)
			}
		}
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
	earlySingletonExposure := meta.IsSingleton() && !meta.IsPrototype() && f.allowCircularReferences && f.singletonComponentRegistry.IsSingletonCurrentlyInCreation(name)
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

func (f *defaultFactory) invokeConstructor(name string, constructor any) (any, error) {
	fnType := reflect.TypeOf(constructor)
	fnValue := reflect.ValueOf(constructor)

	f.logger().Tracef("invoking constructor %s for component '%s' with %d params", fnType.String(), name, fnType.NumIn())

	args := make([]reflect.Value, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		resolved, err := f.resolveConstructorParam(name, i, paramType)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve parameter %d (type %s)", i, paramType)
		}
		args[i] = resolved
	}

	results := fnValue.Call(args)

	if fnType.NumOut() == 2 && !results[1].IsNil() {
		return nil, errors.Wrapf(results[1].Interface().(error), "constructor returned error")
	}
	return results[0].Interface(), nil
}

func (f *defaultFactory) resolveConstructorParam(componentName string, paramIndex int, paramType reflect.Type) (reflect.Value, error) {
	isSlice := paramType.Kind() == reflect.Slice

	var elemType reflect.Type
	if isSlice {
		elemType = paramType.Elem()
	} else {
		elemType = paramType
	}

	var typeOption container.Option
	switch elemType.Kind() {
	case reflect.Ptr:
		typeOption = container.Type(elemType)
	case reflect.Interface:
		typeOption = container.InterfaceType(elemType)
	default:
		return reflect.Value{}, errors.Errorf("unsupported parameter type %s, must be pointer, interface, or slice of them", paramType)
	}

	metas := f.definitionRegistry.GetMetas(typeOption)

	var validMetas []*component_definition.Meta
	for _, m := range metas {
		if m != nil {
			validMetas = append(validMetas, m)
		}
	}

	if len(validMetas) == 0 && !isSlice && elemType.Kind() == reflect.Ptr {
		if resolved, err := f.resolveConfigurationProperties(elemType); err != nil {
			return reflect.Value{}, errors.Wrapf(err, "resolve ConfigurationProperties parameter %d (type %s)", paramIndex, paramType)
		} else if resolved.IsValid() {
			return resolved, nil
		}
	}

	if len(validMetas) == 0 {
		if isSlice {
			return reflect.MakeSlice(paramType, 0, 0), nil
		}
		reason := fmt.Sprintf("no component found for constructor parameter[%d] type %s", paramIndex, paramType)
		return reflect.Value{}, fmt.Errorf("%s\n%s", reason, f.formatDependencyChain(paramType.String(), reason))
	}

	if isSlice {
		slice := reflect.MakeSlice(paramType, len(validMetas), len(validMetas))
		for i, m := range validMetas {
			dep, err := f.doGetComponent(m.Name())
			if err != nil {
				return reflect.Value{}, errors.Wrapf(err, "resolve slice element %d", i)
			}
			slice.Index(i).Set(dep.Value)
		}
		return slice, nil
	}

	selected := component_definition.SelectBestCandidate(validMetas)
	dep, err := f.doGetComponent(selected.Name())
	if err != nil {
		return reflect.Value{}, err
	}
	return dep.Value, nil
}

func (f *defaultFactory) resolveConfigurationProperties(ptrType reflect.Type) (reflect.Value, error) {
	instance := reflect.New(ptrType.Elem()).Interface()
	cp, ok := instance.(definition.ConfigurationProperties)
	if !ok {
		return reflect.Value{}, nil
	}
	prefix := cp.Prefix()
	if f.configure == nil {
		return reflect.Value{}, errors.Errorf("configure is nil, cannot resolve ConfigurationProperties with prefix '%s'", prefix)
	}
	configValue := f.configure.Get(prefix)
	if configValue == nil {
		return reflect.ValueOf(instance), nil
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           instance,
		TagName:          "yaml",
	})
	if err != nil {
		return reflect.Value{}, errors.Wrapf(err, "create decoder for ConfigurationProperties prefix '%s'", prefix)
	}
	if err := decoder.Decode(configValue); err != nil {
		return reflect.Value{}, errors.Wrapf(err, "decode ConfigurationProperties prefix '%s'", prefix)
	}
	return reflect.ValueOf(instance), nil
}

func (f *defaultFactory) logger() syslog.Logger {
	return syslog.Pref("ComponentFactory")
}
