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
	"strings"
)

type defaultFactory struct {
	singletonRegistry                         support.SingletonRegistry
	definitionRegistry                        support.DefinitionRegistry
	singletonComponentRegistry                support.SingletonComponentRegistry
	configure                                 configure.Configure
	definitionRegistryPostProcessors          []processors.DefinitionRegistryPostProcessor
	factoryPostProcessors                     []ComponentFactoryPostProcessor
	postProcessors                            []processors.ComponentPostProcessor
	instantiationAwareComponentPostProcessors []processors.InstantiationAwareComponentPostProcessor
	destructionAwareComponentPostProcessors   []processors.DestructionAwareComponentPostProcessor
	initializedPostProcessors                 []processors.ComponentInitializedPostProcessor
	injectionRules                            []InjectionRule
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
		if p, ok := singleton.(processors.DefinitionRegistryPostProcessor); ok {
			f.definitionRegistryPostProcessors = append(f.definitionRegistryPostProcessors, p)
		}
		if p, ok := singleton.(ComponentFactoryPostProcessor); ok {
			f.factoryPostProcessors = append(f.factoryPostProcessors, p)
		}
		if p, ok := singleton.(processors.ComponentPostProcessor); ok {
			f.postProcessors = append(f.postProcessors, p)
		}
		if p, ok := singleton.(processors.InstantiationAwareComponentPostProcessor); ok {
			f.instantiationAwareComponentPostProcessors = append(f.instantiationAwareComponentPostProcessors, p)
		}
		if p, ok := singleton.(processors.DestructionAwareComponentPostProcessor); ok {
			f.destructionAwareComponentPostProcessors = append(f.destructionAwareComponentPostProcessors, p)
		}
		if p, ok := singleton.(processors.ComponentInitializedPostProcessor); ok {
			f.initializedPostProcessors = append(f.initializedPostProcessors, p)
		}
		singletons[name] = singleton
	}
	for name, singleton := range singletons {
		for _, processor := range f.definitionRegistryPostProcessors {
			err := processor.PostProcessDefinitionRegistry(f.definitionRegistry, singleton, name)
			if err != nil {
				return err
			}
		}
	}
	//syslog.Trace("start populating properties...")
	//err := f.configure.PopulateProperties(metas...)
	//if err != nil {
	//	return fmt.Errorf("populate components properties: %v", err)
	//}
	//syslog.Info("populate properties finished")

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
	names := f.singletonRegistry.GetSingletonNames()
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

	if len(f.instantiationAwareComponentPostProcessors) != 0 {
		if meta == nil {
			return nil, fmt.Errorf("component definition with name '%s' not found", name)
		}
		for _, processor := range f.instantiationAwareComponentPostProcessors {
			component, err := processor.PostProcessBeforeInstantiation(meta, name)
			if err != nil {
				return nil, err
			}
			if component != nil {
				return component, nil
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
	earlySingletonExposure := f.singletonComponentRegistry.IsSingletonCurrentlyInCreation(name)
	if earlySingletonExposure {
		f.logger().Debugf("eagerly caching bean '%s' to allow for resolving potential circular references", name)
		f.singletonComponentRegistry.AddSingletonFactory(name, f.componentSingletonFactoryMethod(meta))
	}

	var exposedComponent = meta
	f.logger().Tracef("start populate component '%s'", name)
	err := f.populateComponent(name, meta)
	if err != nil {
		return nil, err
	}
	f.logger().Debugf("component '%s' population finished", name)

	f.logger().Tracef("start initialize component '%s'", name)
	err = f.initializeComponent(meta)
	if err != nil {
		return nil, err
	}
	f.logger().Debugf("component '%s' initialization finished", name)

	if earlySingletonExposure {
		f.logger().Tracef("try get early singleton reference '%s' for dependents reference version check", name)
		earlySingletonReference, err := f.singletonComponentRegistry.GetSingleton(name, false)
		if err != nil {
			return nil, err
		}
		if earlySingletonReference != nil {
			f.logger().Tracef("early singleton reference for '%s' exists, start check other dependents reference version", name)
			if reflect.DeepEqual(earlySingletonReference, meta) {
				exposedComponent = earlySingletonReference
			} else if dependents := meta.GetDependents(); len(dependents) != 0 {
				for _, dependent := range dependents {
					if f.singletonComponentRegistry.IsSingletonCurrentlyInCreation(dependent) {
						return nil, fmt.Errorf("singleton with name '%s' has been injected into other beans \n%s, but has been wrapped which means that other beans do not use the final version of the bean, please try change component init order.",
							name, strings.Join(dependents, "\n"))
					}
				}
			}
		}
	}
	f.logger().Debugf("do create component '%s' finished", name)
	return exposedComponent, nil
}

func (f *defaultFactory) populateComponent(name string, meta *component_definition.Meta) error {
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
	return nil
}

func (f *defaultFactory) componentSingletonFactoryMethod(m *component_definition.Meta) support.SingletonFactory {
	return support.FuncSingletonFactory(func() (*component_definition.Meta, error) {
		logger := f.logger().Pref("SingletonFactory")
		var wrappedComponent, componentName = m.Raw, m.Name()
		var err error
		logger.Tracef("start to apply post process before component '%s' initialization", componentName)
		wrappedComponent, err = f.applyPostProcessBeforeInitialization(wrappedComponent, componentName)
		if err != nil {
			return nil, err
		}
		if ic, ok := wrappedComponent.(definition.InitializingComponent); ok {
			logger.Tracef("invoking after properties set method for component '%s'", componentName)
			err := ic.AfterPropertiesSet()
			if err != nil {
				return nil, fmt.Errorf("invoking after properties set method for component '%s' error: %s", componentName, err)
			}
		}
		logger.Tracef("start to apply post process after component '%s' initialization", componentName)
		wrappedComponent, err = f.applyPostProcessAfterInitialization(wrappedComponent, componentName)
		if err != nil {
			return nil, err
		}
		if !reflect.DeepEqual(wrappedComponent, m.Raw) {
			logger.Debugf("detected a proxy instance for '%s', return proxy component", componentName)
			m.UseProxy(wrappedComponent)
		}
		return m, nil
	})
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

func (f *defaultFactory) initializeComponent(meta *component_definition.Meta) error {
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
