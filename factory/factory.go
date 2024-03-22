package factory

import (
	"fmt"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/samber/lo"
	"reflect"
	"sort"
	"strings"
)

type defaultFactory struct {
	registry       registry.Registry
	configure      configure.Configure
	injectionRules []InjectionRule
	postProcessors []defination.ComponentPostProcessor
}

func Default() Factory {
	f := &defaultFactory{}
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

func (f *defaultFactory) PrepareSpecialComponents() error {
	cppMetas := f.registry.GetComponents(registry.Interface(new(defination.ComponentPostProcessor)))
	err := f.Initialize(cppMetas...)
	if err != nil {
		return err
	}
	f.postProcessors = make([]defination.ComponentPostProcessor, len(cppMetas))
	for i, pm := range cppMetas {
		syslog.Tracef("register component post processor %s", pm.ID())
		f.postProcessors[i] = pm.Raw.(defination.ComponentPostProcessor)
		f.registry.RemoveComponents(pm.Name)
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

func (f *defaultFactory) Initialize(metas ...*meta.Meta) error {
	for _, m := range metas {
		err := f.initialize(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *defaultFactory) initialize(m *meta.Meta) error {
	syslog.Tracef("start initialize component %s", m.ID())
	if f.registry.IsComponentInited(m.Name) {
		syslog.Tracef("component %s is already init, skip initialize", m.ID())
		return nil
	}
	f.registry.ComponentInited(m.Name)

	syslog.Tracef("start inject dependencies %s", m.ID())
	for _, node := range m.GetComponentNodes() {
		dependencies, err := f.getDependencies(m.ID(), node)
		if err != nil {
			return fmt.Errorf("get dependencies failed: %v", err)
		}
		for _, dependency := range dependencies {
			err := f.initialize(dependency)
			if err != nil {
				return err
			}
		}
		err = node.Inject(dependencies)
		if err != nil {
			return fmt.Errorf("inject dependencies failed: %v", err)
		}
	}

	err := f.doInitialize(m)
	if err != nil {
		return fmt.Errorf("factory initialize component %s failed: %v", m.ID(), err)
	}

	syslog.Tracef("factory initialized component %s", m.ID())
	return nil
}

func (f *defaultFactory) doInitialize(m *meta.Meta) error {
	err := f.applyPostProcessBeforeInitialization(m)
	if err != nil {
		return err
	}
	// init
	if ic, ok := m.Raw.(defination.InitializeComponent); ok {
		syslog.Tracef("component %s is InitializeComponent, start do init", m.ID())
		err := ic.Init()
		if err != nil {
			return fmt.Errorf("component %s inited failed: %s", m.ID(), err)
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
		err := processor.PostProcessBeforeInitialization(m.Raw, m.Name)
		if err != nil {
			return fmt.Errorf("post processor: %T process before %s init error: %v", processor, m.ID(), err)
		}
	}
	return nil
}

func (f *defaultFactory) applyPostProcessAfterInitialization(m *meta.Meta) error {
	for _, processor := range f.postProcessors {
		err := processor.PostProcessAfterInitialization(m.Raw, m.Name)
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
