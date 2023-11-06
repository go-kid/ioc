package injector

import (
	"fmt"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/util/list"
	"log"
)

type InjectProcessor interface {
	RuleName() string
	Filter(d *meta.Node) bool
	Inject(r registry.Registry, d *meta.Node) error
}

var modifyInjectors []InjectProcessor

func AddModifyInjectors(injectors []InjectProcessor) {
	modifyInjectors = append(modifyInjectors, injectors...)
}

var injectors = []InjectProcessor{
	new(specifyInjector),
	new(unSpecifyPtrInjector),
	new(unSpecifyInterfaceInjector),
	new(unSpecifyInterfaceSliceInjector),
	new(customizedPtrInjector),
	new(customizedInterfaceInjector),
	new(customizedInterfaceSliceInjector),
}

func DependencyInject(r registry.Registry, id string, dependencies []*meta.Node) error {
	for _, dependency := range dependencies {
		err := injectDependency(append(modifyInjectors, injectors...), r, id, dependency)
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

func injectDependency(injectors []InjectProcessor, r registry.Registry, metaID string, d *meta.Node) error {
	i, find := list.NewList(injectors).FindBy(func(i int) bool {
		return injectors[i].Filter(d)
	})
	if !find {
		return fmt.Errorf(diErrOutput, "nil", metaID, d.Id(), "injection condition not found")
	}
	inj := injectors[i]
	defer func() {
		if err := recover(); err != nil {
			log.Panicf(diErrOutput, inj.RuleName(), metaID, d.Id(), err)
		}
	}()
	err := inj.Inject(r, d)
	if err != nil {
		return fmt.Errorf(diErrOutput, inj.RuleName(), metaID, d.Id(), err)
	}
	return nil
}
