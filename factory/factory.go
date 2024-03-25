package factory

import (
	"fmt"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/ioc/util/sync2"
	"github.com/samber/lo"
	"reflect"
	"sort"
	"strings"
)

type defaultFactory struct {
	container
	registry                   registry.Registry
	configure                  configure.Configure
	injectionRules             []InjectionRule
	postInitializingProcessors []defination.ComponentPostInitializingProcessor
	postProcessors             []defination.ComponentPostProcessor
}

type FactoryMethod func() (*meta.Meta, error)

func Default() Factory {
	f := &defaultFactory{
		container: container{
			metaMaps:              sync2.New[string, *meta.Meta](),
			singletonObjects:      sync2.New[string, *meta.Meta](),
			earlySingletonObjects: sync2.New[string, *meta.Meta](),
			singletonFactories:    sync2.New[string, FactoryMethod](),
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
	f.registry.GetComponents()
	cppMetas := f.registry.GetInternalComponents()
	for _, cppMeta := range cppMetas {
		switch cpp := cppMeta.Raw; cpp.(type) {
		case defination.ComponentPostInitializingProcessor:
			f.postInitializingProcessors = append(f.postInitializingProcessors, cpp.(defination.ComponentPostInitializingProcessor))
		case defination.ComponentPostProcessor:
			f.postProcessors = append(f.postProcessors, cpp.(defination.ComponentPostProcessor))
		}
	}
	return nil
}

func (f *defaultFactory) SetRegistry(r registry.Registry) {
	f.registry = r
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

func (f *defaultFactory) Initialize() error {
	for _, m := range f.registry.GetInternalComponents() {
		_, err := f.getOrInitiateComponentMeta(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *defaultFactory) GetComponent(name string) (*meta.Meta, error) {
	m := f.registry.GetInternalComponentByName(name)
	return f.getOrInitiateComponentMeta(m)
}

func (f *defaultFactory) getOrInitiateComponentMeta(m *meta.Meta) (*meta.Meta, error) {
	// get component from inited components cache
	if f.registry.IsComponentInited(m.Name) {
		return f.registry.GetComponentByName(m.Name), nil
	}
	// get component from early export components cache
	if earlyComponent, ok := f.registry.GetEarlyExportComponent(m.Name); ok {
		return earlyComponent, nil
	}
	// get component from singleton component factory cache
	if factoryMethod, ok := f.registry.GetSingletonFactoryMethod(m.Name); ok {
		createdComponent, err := factoryMethod()
		if err != nil {
			return nil, err
		}
		f.registry.EarlyExportComponent(createdComponent)
		return createdComponent, nil
	}

	// set to singleton component factory cache
	f.registry.SetSingletonFactoryMethod(m.Name, func() (*meta.Meta, error) {
		return f.initializingComponent(m)
	})
	// inject dependencies
	for _, node := range m.GetComponentNodes() {
		dependencies, err := f.getDependencies(m.ID(), node)
		if err != nil {
			return nil, err
		}
		var wrappedDependencies = make([]*meta.Meta, len(dependencies))
		for i, dependency := range dependencies {
			wrappedDependency, err := f.getOrInitiateComponentMeta(dependency)
			if err != nil {
				return nil, err
			}
			wrappedDependencies[i] = wrappedDependency
		}
		err = node.Inject(wrappedDependencies)
		if err != nil {
			return nil, fmt.Errorf("inject dependencies failed: %v", err)
		}
	}
	err := f.doInitialize(m)
	if err != nil {
		return nil, err
	}
	err = f.registry.ComponentInited(m.Name)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (f *defaultFactory) initializingComponent(m *meta.Meta) (*meta.Meta, error) {
	for _, processor := range f.postInitializingProcessors {
		wrappedComponent, err := processor.PostProcessBeforeInitializing(m)
		if err != nil {
			return nil, err
		}
		if wrappedComponent == nil {
			return m, nil
		}
		m = wrappedComponent
	}
	if ic, ok := m.Raw.(defination.InitializingComponent); ok {
		syslog.Tracef("initializing component %s do initialization", m.ID())
		err := ic.Initializing()
		if err != nil {
			return nil, fmt.Errorf("initializing component %s initialization failed: %s", m.ID(), err)
		}
	}
	for _, processor := range f.postInitializingProcessors {
		wrappedComponent, err := processor.PostProcessAfterInitializing(m)
		if err != nil {
			return nil, err
		}
		if wrappedComponent == nil {
			return m, nil
		}
		m = wrappedComponent
	}
	return m, nil
}

func (f *defaultFactory) doInitialize(m *meta.Meta) error {
	err := f.applyPostProcessBeforeInitialization(m)
	if err != nil {
		return err
	}
	// init
	if ic, ok := m.Raw.(defination.InitializeComponent); ok {
		syslog.Tracef("initialize component %s do init", m.ID())
		err := ic.Init()
		if err != nil {
			return fmt.Errorf("initialize component %s do init failed: %s", m.ID(), err)
		}
	}
	err = f.applyPostProcessAfterInitialization(m)
	if err != nil {
		return err
	}
	return nil
}

func (f *defaultFactory) applyPostProcessBeforeInitialization(m *meta.Meta) error {
	for _, processor := range f.postProcessors {
		err := processor.PostProcessBeforeInitialization(m.Raw)
		if err != nil {
			return fmt.Errorf("post processor: %T process before %s init error: %v", processor, m.ID(), err)
		}
	}
	return nil
}

func (f *defaultFactory) applyPostProcessAfterInitialization(m *meta.Meta) error {
	for _, processor := range f.postProcessors {
		err := processor.PostProcessAfterInitialization(m.Raw)
		if err != nil {
			return fmt.Errorf("post processor: %T process after %s init error: %v", processor, m.ID(), err)
		}
	}
	return nil
}

const diErrOutput = "DI report error by processor: %s\n" +
	"caused instance: %s\n" +
	"caused field: %s\n" +
	"caused by: %v\n"

func (f *defaultFactory) getDependencies(metaID string, d *meta.Node) ([]*meta.Meta, error) {
	inj, find := lo.Find(f.injectionRules, func(item InjectionRule) bool {
		return item.Condition(d)
	})
	if !find {
		return nil, fmt.Errorf(diErrOutput, "nil", metaID, d.ID(), "inject condition not found")
	}
	defer func() {
		if err := recover(); err != nil {
			syslog.Panicf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
		}
	}()
	candidates, err := inj.Candidates(f.registry, d)
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
			if d.Args().Has(meta.ArgRequired, "true") {
				return nil, fmt.Errorf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
			}
			return nil, nil
		}
		return nil, fmt.Errorf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
	}
	return candidates, nil
}

var (
	primaryInterface = new(defination.WirePrimary)
)

func filterDependencies(n *meta.Node, metas []*meta.Meta) ([]*meta.Meta, error) {
	//remove nil meta
	result := filter(metas, func(m *meta.Meta) bool {
		return m != nil
	})
	if len(result) == 0 {
		return nil, fmt.Errorf("%s not found available components", n.ID())
	}
	//remove self-inject
	result = filter(result, func(m *meta.Meta) bool {
		return m.ID() != n.Holder.Meta.ID()
	})
	if len(result) == 0 {
		var embedSb = strings.Builder{}
		_ = n.Holder.Walk(func(source *meta.Holder) error {
			embedSb.WriteString("\n depended on " + source.ID())
			return nil
		})
		return nil, fmt.Errorf("field %s %s: self inject not allowed", n.ID(), embedSb.String())
	}
	//filter qualifier
	qualifierName, isQualifier := n.Args().Find(meta.ArgQualifier)
	if isQualifier {
		result = filter(result, func(m *meta.Meta) bool {
			qualifier, ok := m.Raw.(defination.WireQualifier)
			return ok && n.Args().Has(meta.ArgQualifier, qualifier.Qualifier())
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
			if !m.IsAlias {
				candidate = m
			}
		}
		result = []*meta.Meta{candidate}
	}
	return result, nil
}

func filter(metas []*meta.Meta, f func(m *meta.Meta) bool) []*meta.Meta {
	var result = make([]*meta.Meta, 0, len(metas))
	for _, m := range metas {
		if f(m) {
			result = append(result, m)
		}
	}
	return result
}
