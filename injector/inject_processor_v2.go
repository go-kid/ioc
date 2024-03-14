package injector

import (
	"fmt"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/samber/lo"
	"sort"
)

type injector struct {
	injectors []InjectProcessor
}

func Default() Injector {
	ij := &injector{}
	ij.addInjectors([]InjectProcessor{
		new(specifyInjector),
		new(unSpecifyPtrInjector),
		new(unSpecifyPtrSliceInjector),
		new(unSpecifyInterfaceInjector),
		new(unSpecifyInterfaceSliceInjector),
		new(customizedPtrInjector),
		new(customizedInterfaceInjector),
		new(customizedInterfaceSliceInjector),
	})
	return ij
}

func (s *injector) addInjectors(ips []InjectProcessor) {
	s.injectors = append(s.injectors, ips...)
	sort.Slice(s.injectors, func(i, j int) bool {
		return s.injectors[i].Priority() < s.injectors[j].Priority()
	})
}

func (s *injector) AddCustomizedInjectors(ips ...InjectProcessor) {
	s.addInjectors(ips)
}

func (s *injector) DependencyInject(r registry.Registry, id string, dependencies []*meta.Node) error {
	for _, dependency := range dependencies {
		err := s.injectDependency(r, id, dependency)
		if err != nil {
			return err
		}
	}
	return nil
}

const diErrOutput = "DI report error by processor: %s\n" +
	"caused instance: %s\n" +
	"caused field: %s\n" +
	"caused by: %v\n"

func (s *injector) injectDependency(r registry.Registry, metaID string, d *meta.Node) error {
	inj, find := lo.Find(s.injectors, func(item InjectProcessor) bool {
		return item.Filter(d)
	})
	if !find {
		return fmt.Errorf(diErrOutput, "nil", metaID, d.ID(), "inject condition not found")
	}
	defer func() {
		if err := recover(); err != nil {
			syslog.Panicf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
		}
	}()
	err := inj.Inject(r, d)
	if err != nil {
		return fmt.Errorf(diErrOutput, inj.RuleName(), metaID, d.ID(), err)
	}
	return nil
}
