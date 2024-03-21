package factory

import (
	"fmt"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/samber/lo"
	"regexp"
	"sort"
)

type defaultFactory struct {
	expReg         *regexp.Regexp
	registry       registry.Registry
	configure      configure.Configure
	injectionRules []InjectionRule
	postProcessors []defination.ComponentPostProcessor
}

func Default() Factory {
	f := &defaultFactory{
		expReg: regexp.MustCompile("\\#\\{[\\d\\w]+(\\.[\\d\\w]+)*(:[\\d\\w]*)?\\}"),
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

	if nodes := m.GetComponentNodes(); len(nodes) > 0 {
		syslog.Tracef("start inject dependencies %s", m.ID())
		for _, node := range nodes {
			err := f.injectDependency(m.ID(), node)
			if err != nil {
				return fmt.Errorf("inject dependencies failed: %v", err)
			}
		}

		for _, node := range nodes {
			for _, sub := range node.Injects {
				err := f.initialize(sub)
				if err != nil {
					return err
				}
			}
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
	// before process
	for _, processor := range f.postProcessors {
		err := processor.PostProcessBeforeInitialization(m.Raw)
		if err != nil {
			return fmt.Errorf("post processor: %T process before %s init error: %v", processor, m.ID(), err)
		}
	}
	// init
	if ic, ok := m.Raw.(defination.InitializeComponent); ok {
		syslog.Tracef("component %s is InitializeComponent, start do init", m.ID())
		err := ic.Init()
		if err != nil {
			return fmt.Errorf("component %s inited failed: %s", m.ID(), err)
		}
	}

	// after process
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

func (f *defaultFactory) injectDependency(metaID string, d *meta.Node) error {
	inj, find := lo.Find(f.injectionRules, func(item InjectionRule) bool {
		return item.Condition(d)
	})
	if !find {
		return fmt.Errorf(diErrOutput, "nil", metaID, d.ID(), "inject condition not found")
	}
	defer func() {
		if err := recover(); err != nil {
			syslog.Panicf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
		}
	}()
	candidates, err := inj.Candidates(f.registry, d)
	if err != nil {
		return fmt.Errorf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
	}
	err = d.Inject(candidates)
	if err != nil {
		return fmt.Errorf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
	}
	return nil
}
