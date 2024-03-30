package factory

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/factory/processors/definition_registry_post_processors"
	"github.com/go-kid/ioc/factory/support"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"reflect"
	"sort"
)

type defaultFactory struct {
	singletonRegistry                           support.SingletonRegistry
	definitionRegistry                          support.DefinitionRegistry
	singletonComponentRegistry                  support.SingletonComponentRegistry
	configure                                   configure.Configure
	definitionRegistryPostProcessors            []processors.DefinitionRegistryPostProcessor
	factoryPostProcessors                       []ComponentFactoryPostProcessor
	postProcessors                              []processors.ComponentPostProcessor
	hasInstantiationAwareComponentPostProcessor bool
	hasDestructionAwareComponentPostProcessor   bool
	initializedPostProcessors                   []processors.ComponentInitializedPostProcessor
	injectionRules                              []InjectionRule
	allowCircularReferences                     bool
}

func Default() Factory {
	f := &defaultFactory{
		definitionRegistry:         support.DefaultDefinitionRegistry(),
		singletonComponentRegistry: support.DefaultSingletonComponentRegistry(),
		definitionRegistryPostProcessors: []processors.DefinitionRegistryPostProcessor{
			&definition_registry_post_processors.PropTagScanProcessor{},
			&definition_registry_post_processors.ValueTagScanProcessor{},
			&definition_registry_post_processors.WireTagScanProcessor{},
		},
		allowCircularReferences: true,
	}
	f.AddInjectionRules(
		new(specifyInjector),
		new(unSpecifyPtrInjector),
		new(unSpecifyPtrSliceInjector),
		new(unSpecifyInterfaceInjector),
		new(unSpecifyInterfaceSliceInjector),
		new(customizedPtrInjector),
		new(customizedInterfaceInjector),
		new(customizedInterfaceSliceInjector),
	)
	return f
}

func (f *defaultFactory) PrepareComponents() error {
	singletonNames := f.singletonRegistry.GetSingletonNames()
	var singletons = make(map[string]any, len(singletonNames))
	for _, name := range singletonNames {
		singleton, err := f.singletonRegistry.GetSingleton(name)
		if err != nil {
			return err
		}
		switch singleton.(type) {
		case processors.ComponentPostProcessor:
			if _, ok := singleton.(processors.InstantiationAwareComponentPostProcessor); ok {
				f.hasInstantiationAwareComponentPostProcessor = true
			}
			if _, ok := singleton.(processors.DestructionAwareComponentPostProcessor); ok {
				f.hasDestructionAwareComponentPostProcessor = true
			}
			f.postProcessors = append(f.postProcessors, singleton.(processors.ComponentPostProcessor))
		case processors.DefinitionRegistryPostProcessor:
			f.definitionRegistryPostProcessors = append(f.definitionRegistryPostProcessors, singleton.(processors.DefinitionRegistryPostProcessor))
		case ComponentFactoryPostProcessor:
			f.factoryPostProcessors = append(f.factoryPostProcessors, singleton.(ComponentFactoryPostProcessor))
		case processors.ComponentInitializedPostProcessor:
			f.initializedPostProcessors = append(f.initializedPostProcessors, singleton.(processors.ComponentInitializedPostProcessor))
		default:
			singletons[name] = singleton
		}
	}

	for name, singleton := range singletons {
		err := f.postProcessDefinitionRegistry(name, singleton)
		if err != nil {
			return err
		}
	}

	if len(f.factoryPostProcessors) != 0 {
		for _, fp := range f.factoryPostProcessors {
			err := fp.PostProcessComponentFactory(f)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *defaultFactory) postProcessDefinitionRegistry(name string, component any) error {
	for _, processor := range f.definitionRegistryPostProcessors {
		err := processor.PostProcessDefinitionRegistry(f.definitionRegistry, component, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *defaultFactory) SetRegistry(r support.SingletonRegistry) {
	f.singletonRegistry = r
}

func (f *defaultFactory) SetConfigure(c configure.Configure) {
	f.configure = c
}

func (f *defaultFactory) AddInjectionRules(rules ...InjectionRule) {
	f.injectionRules = append(f.injectionRules, rules...)
	sort.Slice(f.injectionRules, func(i, j int) bool {
		return f.injectionRules[i].Priority() < f.injectionRules[j].Priority()
	})
}

func (f *defaultFactory) Refresh() error {

	var names []string

	for _, meta := range f.definitionRegistry.GetMetas() {
		switch meta.Raw.(type) {
		case definition.LazyInitComponent:
			continue
		default:
			names = append(names, meta.Name())
		}
	}

	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	for _, name := range names {
		f.logger().Tracef("refresh component with name '%s'", name)
		_, err := f.doGetComponent(name)
		if err != nil {
			return fmt.Errorf("refresh component with name '%s' failed: %v", name, err)
		}
	}
	return nil
}

func (f *defaultFactory) GetComponents(opts ...support.Option) ([]any, error) {
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
		support.FuncSingletonFactory(func() (*component_definition.Meta, error) {
			return f.createComponent(name)
		}))
	if err != nil {
		return nil, err
	}
	return sharedInstance, nil
}

func (f *defaultFactory) createComponent(name string) (*component_definition.Meta, error) {
	meta := f.definitionRegistry.GetMetaByName(name)

	if f.hasInstantiationAwareComponentPostProcessor {
		if meta == nil {
			return nil, fmt.Errorf("component definition with name '%s' not found", name)
		}
		for _, processor := range f.postProcessors {
			if ipb, ok := processor.(processors.InstantiationAwareComponentPostProcessor); ok {
				component, err := ipb.PostProcessBeforeInstantiation(meta, name)
				if err != nil {
					return nil, err
				}
				if component != nil {
					return component, nil
				}
			}
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
		f.singletonComponentRegistry.AddSingletonFactory(name, support.FuncSingletonFactory(func() (*component_definition.Meta, error) {
			return f.getEarlyBeanReference(name, meta)
		}))
	}

	var exposedComponent = meta

	err := f.populateComponent(name, meta)
	if err != nil {
		return nil, err
	}

	exposedComponent, err = f.initializeComponent(name, meta)
	if err != nil {
		return nil, err
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
					return nil, fmt.Errorf("singleton with name '%s' has been injected into other components \n%s, but has been wrapped which means that other beans do not use the final version of the bean, please try change component init order.",
						name, actualDependents)
				}
			}
		}
	}
	err = f.processInitializedComponentInitialization(meta)
	if err != nil {
		return nil, err
	}
	f.logger().Debugf("do create component '%s' finished", name)
	return exposedComponent, nil
}

func (f *defaultFactory) populateComponent(name string, meta *component_definition.Meta) error {
	f.logger().Tracef("start populate component '%s'", name)
	if len(meta.GetConfigurationNodes()) != 0 {
		f.logger().Tracef("populate properties for '%s'", name)
		err := f.configure.PopulateProperties(meta)
		if err != nil {
			return err
		}
	}
	if nodes := meta.GetComponentNodes(); len(nodes) > 0 {
		f.logger().Tracef("inject dependencies for '%s'", name)
		for _, node := range meta.GetComponentNodes() {
			f.logger().Tracef("get dependencies for %s", node.ID())
			dependencies, err := f.getDependencies(meta.ID(), node)
			if err != nil {
				return err
			}
			var injects []*component_definition.Meta
			for _, dependency := range dependencies {
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
		f.logger().Tracef("finished inject dependencies for '%s'", name)
	}
	f.logger().Debugf("component '%s' population finished", name)
	return nil
}

func (f *defaultFactory) getEarlyBeanReference(name string, m *component_definition.Meta) (*component_definition.Meta, error) {
	var exposedComponent = m.Raw
	var err error
	if f.hasInstantiationAwareComponentPostProcessor {
		for _, processor := range f.postProcessors {
			if ibp, ok := processor.(processors.SmartInstantiationAwareBeanPostProcessor); ok {
				exposedComponent, err = ibp.GetEarlyBeanReference(exposedComponent, name)
				if err != nil {
					return nil, err
				}
			}
		}
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

func (f *defaultFactory) initializeComponent(name string, m *component_definition.Meta) (*component_definition.Meta, error) {
	f.logger().Tracef("start initialize component '%s'", name)
	logger := f.logger()
	var wrappedComponent = m.Raw
	var err error
	logger.Tracef("start to apply post process before component '%s' initialization", name)
	wrappedComponent, err = f.applyPostProcessBeforeInitialization(wrappedComponent, name)
	if err != nil {
		return nil, err
	}
	err = f.invokeInitMethods(name, wrappedComponent)
	if err != nil {
		return nil, err
	}
	logger.Tracef("start to apply post process after component '%s' initialization", name)
	wrappedComponent, err = f.applyPostProcessAfterInitialization(wrappedComponent, name)
	if err != nil {
		return nil, err
	}
	if wrappedComponent != m.Raw {
		//m = component_definition.NewMeta(wrappedComponent)
		//m.SetName(name)
		//m.UseProxy(wrappedComponent)
		m, err = f.genProxyComponent(m, name, wrappedComponent)
		if err != nil {
			return nil, err
		}
		logger.Debugf("component '%s' initialization finished, detected a new instance '%s' proxy, return proxy component", name, m.Type.String())
	} else {
		f.logger().Debugf("component '%s' initialization finished", name)
	}
	return m, nil
}

func (f *defaultFactory) invokeInitMethods(name string, component any) error {
	if ic, ok := component.(definition.InitializingComponent); ok {
		f.logger().Tracef("invoking afterPropertiesSet() method for component '%s'", name)
		err := ic.AfterPropertiesSet()
		if err != nil {
			return fmt.Errorf("invoking afterPropertiesSet() method for component '%s' error: %s", name, err)
		}
	}
	return nil
}

func (f *defaultFactory) applyPostProcessBeforeInitialization(c any, name string) (any, error) {
	var (
		result = c
		err    error
	)
	var current any
	for _, processor := range f.postProcessors {
		current, err = processor.PostProcessBeforeInitialization(result, name)
		if err != nil {
			return nil, fmt.Errorf("component post processor %s apply post process before initialization error: %v", reflectx.Id(processor), err)
		}
		if current == nil {
			return result, nil
		}
		result = current
	}
	return result, nil
}

func (f *defaultFactory) applyPostProcessAfterInitialization(c any, name string) (any, error) {
	var (
		result = c
		err    error
	)
	var current any
	for _, processor := range f.postProcessors {
		current, err = processor.PostProcessAfterInitialization(result, name)
		if err != nil {
			return nil, fmt.Errorf("component post processor %s apply post process after initialization error: %v", reflectx.Id(processor), err)
		}
		if current == nil {
			return result, nil
		}
		result = current
	}
	return result, nil
}

func (f *defaultFactory) processInitializedComponentInitialization(meta *component_definition.Meta) error {
	f.logger().Tracef("post process before initialize component '%s'", meta.Name())
	for _, processor := range f.initializedPostProcessors {
		err := processor.PostProcessBeforeInitialization(meta.Raw)
		if err != nil {
			return err
		}
	}
	if c, ok := meta.Raw.(definition.InitializeComponent); ok {
		f.logger().Tracef("invoking init method for component '%s'", meta.Name())
		err := c.Init()
		if err != nil {
			return err
		}
	}
	f.logger().Tracef("post process after initialize component '%s'", meta.Name())
	for _, processor := range f.initializedPostProcessors {
		err := processor.PostProcessAfterInitialization(meta.Raw)
		if err != nil {
			return err
		}
	}
	f.logger().Debugf("initialize component '%s' finished", meta.Name())
	return nil
}

const diErrOutput = "DI report error by processor: %s\n" +
	"caused instance: %s\n" +
	"caused field: %s\n" +
	"caused by: %v\n"

func (f *defaultFactory) getDependencies(metaID string, d *component_definition.Node) ([]*component_definition.Meta, error) {
	inj, find := lo.Find(f.injectionRules, func(item InjectionRule) bool {
		return item.Condition(d)
	})
	if !find {
		return nil, fmt.Errorf(diErrOutput, "nil", metaID, d.ID(), "inject condition not found")
	}
	defer func() {
		if err := recover(); err != nil {
			f.logger().Panicf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
		}
	}()
	candidates, err := inj.Candidates(f.definitionRegistry, d)
	if err != nil {
		return nil, fmt.Errorf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
	}
	//err = d.Inject(candidates)
	//if err != nil {
	//	return fmt.Errorf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
	//}
	candidates, err = filterDependencies(d, candidates)
	if err != nil {
		if len(candidates) == 0 {
			if d.Args().Has(component_definition.ArgRequired, "true") {
				return nil, fmt.Errorf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
			}
			return nil, nil
		}
		return nil, fmt.Errorf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
	}
	return candidates, nil
}

var (
	primaryInterface = new(definition.WirePrimary)
)

func filterDependencies(n *component_definition.Node, metas []*component_definition.Meta) ([]*component_definition.Meta, error) {
	//remove nil meta
	result := filter(metas, func(m *component_definition.Meta) bool {
		return m != nil
	})
	if len(result) == 0 {
		return nil, fmt.Errorf("%s not found available components", n.ID())
	}
	//filter qualifier
	qualifierName, isQualifier := n.Args().Find(component_definition.ArgQualifier)
	if isQualifier {
		result = filter(result, func(m *component_definition.Meta) bool {
			qualifier, ok := m.Raw.(definition.WireQualifier)
			return ok && n.Args().Has(component_definition.ArgQualifier, qualifier.Qualifier())
		})
		if len(result) == 0 {
			return nil, fmt.Errorf("field %s: no component found for qualifier %s", n.ID(), qualifierName)
		}
	}

	//filter primary for single type
	if len(result) > 1 && n.Type.Kind() != reflect.Slice && n.Type.Kind() != reflect.Array {
		var candidate = result[0]

		for _, m := range result {
			//Primary interface first
			if reflectx.IsTypeImplement(m.Type, primaryInterface) {
				candidate = m
				break
			}
			//non naming component is preferred in multiple candidates
			if !m.IsAlias() {
				candidate = m
			}
		}
		result = []*component_definition.Meta{candidate}
	}
	return result, nil
}

func filter(metas []*component_definition.Meta, f func(m *component_definition.Meta) bool) []*component_definition.Meta {
	var result = make([]*component_definition.Meta, 0, len(metas))
	for _, m := range metas {
		if f(m) {
			result = append(result, m)
		}
	}
	return result
}

func (f *defaultFactory) GetDefinitionRegistry() support.DefinitionRegistry {
	return f.definitionRegistry
}

func (f *defaultFactory) logger() syslog.Logger {
	return syslog.GetLogger().Pref("ComponentFactory")
}
