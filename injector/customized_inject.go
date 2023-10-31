package injector

import (
	"fmt"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"reflect"
)

var customizedInjectors = []injectProcessor{
	new(customizedPtrInjector),
	new(customizedInterfaceInjector),
	new(customizedInterfaceSliceInjector),
}

func CustomizedInject(r registry.Registry, id string, customized []*meta.Node) error {
	for _, node := range customized {
		err := injectDependency(customizedInjectors, r, id, node)
		if err != nil {
			return err
		}
	}
	return nil
}

type customizedPtrInjector struct {
}

func (c *customizedPtrInjector) Filter(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Ptr
}

func (c *customizedPtrInjector) Inject(r registry.Registry, d *meta.Node) error {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.Type(d.Type))
	if len(metas) < 1 {
		return fmt.Errorf("no instance found for customized tag [%s] with type: %s", d.Tag, d.Type)
	}
	d.Inject(metas[0])
	return nil
}

type customizedInterfaceInjector struct {
}

func (c *customizedInterfaceInjector) Filter(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Interface
}

func (c *customizedInterfaceInjector) Inject(r registry.Registry, d *meta.Node) error {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.InterfaceType(d.Type))
	if len(metas) < 1 {
		return fmt.Errorf("no instance found for customized tag [%s] implement the interface: %s", d.Tag, d.Type)
	}
	d.Inject(metas[0])
	return nil
}

type customizedInterfaceSliceInjector struct {
}

func (s *customizedInterfaceSliceInjector) Filter(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Slice && d.Type.Elem().Kind() == reflect.Interface //ruleSliceInterface
}

func (s *customizedInterfaceSliceInjector) Inject(r registry.Registry, d *meta.Node) error {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.InterfaceType(d.Type.Elem()))
	d.Inject(metas...)
	return nil
}
