package factory

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"reflect"
	"sort"
	"strings"
)

type defaultFactory struct {
	definitionRegistry DefinitionRegistry
	scanner            scanner.Scanner
	singletonRegistry  registry.SingletonRegistry
	configure          configure.Configure
	injectionRules     []InjectionRule
	postProcessors     []definition.ComponentPostProcessor
}

func (f *defaultFactory) GetComponents(opts ...Option) []any {
	var results []any
	for _, meta := range f.definitionRegistry.GetComponentDefinitions(opts...) {
		results = append(results, meta.Raw)
	}
	return results
}

func Default() Factory {
	f := &defaultFactory{
		definitionRegistry: DefaultDefinitionRegistry(),
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
	var metas []*component_definition.Meta
	for _, name := range singletonNames {
		singleton, err := f.singletonRegistry.GetSingleton(name)
		if err != nil {
			return err
		}
		meta := f.scanner.ScanComponent(singleton)
		metas = append(metas, meta)
	}
	syslog.Trace("start populating properties...")
	err := f.configure.PopulateProperties(metas...)
	if err != nil {
		return fmt.Errorf("populate components properties: %v", err)
	}
	syslog.Info("populate properties finished")
	for _, m := range metas {
		switch cpp := m.Raw; cpp.(type) {
		case definition.ComponentPostProcessor:
			f.postProcessors = append(f.postProcessors, cpp.(definition.ComponentPostProcessor))
		default:
			f.definitionRegistry.RegisterMeta(m)
		}
	}
	return nil
}

func (f *defaultFactory) SetRegistry(r registry.SingletonRegistry) {
	f.singletonRegistry = r
}

func (f *defaultFactory) SetConfigure(c configure.Configure) {
	f.configure = c
}

func (f *defaultFactory) SetScanner(sc scanner.Scanner) {
	f.scanner = sc
}

func (f *defaultFactory) AddInjectionRules(rules ...InjectionRule) {
	f.injectionRules = append(f.injectionRules, rules...)
	sort.Slice(f.injectionRules, func(i, j int) bool {
		return f.injectionRules[i].Priority() < f.injectionRules[j].Priority()
	})
}

func (f *defaultFactory) Initialize() error {
	for _, s := range f.singletonRegistry.GetSingletonNames() {
		_, err := f.GetComponent(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *defaultFactory) GetComponentByName(name string) (any, error) {
	meta, err := f.definitionRegistry.GetComponent(name)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		meta, err = f.getSingleton(name)
		if err != nil {
			return nil, err
		}
	}
	return meta.Raw, nil
}

func (f *defaultFactory) GetComponent(name string) (*component_definition.Meta, error) {
	meta, err := f.definitionRegistry.GetComponent(name)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		meta, err = f.getSingleton(name)
		if err != nil {
			return nil, err
		}
	}
	return meta, nil
}

func (f *defaultFactory) getSingleton(name string) (*component_definition.Meta, error) {
	f.definitionRegistry.BeforeSingletonCreation(name)
	component, err := f.createComponent(name)
	if err != nil {
		return nil, err
	}
	return component, nil
}

func (f *defaultFactory) createComponent(name string) (*component_definition.Meta, error) {
	var result *component_definition.Meta
	if f.definitionRegistry.IsSingletonCurrentlyInCreation(name) {
		meta := f.definitionRegistry.GetMetaByName(name)
		// set to singleton component factory cache
		f.definitionRegistry.AddSingletonFactory(name, f.componentSingletonFactoryMethod(meta))
		syslog.Tracef("definition registry set singleton factory for %s", name)
		err := f.populateComponent(meta)
		if err != nil {
			return nil, err
		}
		result = meta
	}
	return result, nil
}

func (f *defaultFactory) populateComponent(meta *component_definition.Meta) error {
	err := f.configure.PopulateProperties(meta)
	if err != nil {
		return err
	}
	for _, node := range meta.GetComponentNodes() {
		dependencies, err := f.getDependencies(meta.ID(), node)
		if err != nil {
			return err
		}
		var injects []*component_definition.Meta
		for _, dependency := range dependencies {
			component, err := f.GetComponent(dependency.Name)
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
	f.definitionRegistry.ComponentInitialized(meta)
	return nil
}

func (f *defaultFactory) componentSingletonFactoryMethod(m *component_definition.Meta) SingletonFactory {
	return FuncSingletonFactory(func() (*component_definition.Meta, error) {
		var wrappedComponent, componentName = m.Raw, m.Name
		var err error
		wrappedComponent, err = f.applyPostProcessBeforeInitialization(wrappedComponent, componentName)
		if err != nil {
			return nil, err
		}
		if ic, ok := wrappedComponent.(definition.InitializingComponent); ok {
			syslog.Tracef("initializing component %s do initialization", m.ID())
			err := ic.AfterPropertiesSet()
			if err != nil {
				return nil, fmt.Errorf("initializing component %s initialization failed: %s", m.ID(), err)
			}
		}
		wrappedComponent, err = f.applyPostProcessAfterInitialization(wrappedComponent, componentName)
		if err != nil {
			return nil, err
		}
		return f.scanner.ScanComponent(wrappedComponent), nil
	})
}

//func (f *defaultFactory) doInitialize(m *meta.Meta) error {
//	err := f.applyPostProcessBeforeInitialization(m)
//	if err != nil {
//		return err
//	}
//	// init
//	if ic, ok := m.Raw.(defination.InitializeComponent); ok {
//		syslog.Tracef("initialize component %s do init", m.ID())
//		err := ic.Init()
//		if err != nil {
//			return fmt.Errorf("initialize component %s do init failed: %s", m.ID(), err)
//		}
//	}
//	err = f.applyPostProcessAfterInitialization(m)
//	if err != nil {
//		return err
//	}
//	return nil
//}

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
			syslog.Panicf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
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
	//remove self-inject
	result = filter(result, func(m *component_definition.Meta) bool {
		return m.ID() != n.Holder.Meta.ID()
	})
	if len(result) == 0 {
		var embedSb = strings.Builder{}
		_ = n.Holder.Walk(func(source *component_definition.Holder) error {
			embedSb.WriteString("\n depended on " + source.ID())
			return nil
		})
		return nil, fmt.Errorf("field %s %s: self inject not allowed", n.ID(), embedSb.String())
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
			if !m.IsAlias {
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
