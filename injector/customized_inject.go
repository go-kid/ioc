package injector

import (
	"fmt"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"reflect"
)

type customizedPtrInjector struct {
}

func (c *customizedPtrInjector) RuleName() string {
	return "Customized_Pointer"
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

func (c *customizedInterfaceInjector) RuleName() string {
	return "Customized_Interface"
}

func (c *customizedInterfaceInjector) Filter(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Interface
}

func (c *customizedInterfaceInjector) Inject(r registry.Registry, d *meta.Node) error {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.InterfaceType(d.Type))
	if len(metas) < 1 {
		return fmt.Errorf("none instance found for customized tag [%s] implement the interface: %s", d.Tag, d.Type)
	}
	d.Inject(metas[0])
	return nil
}

type customizedInterfaceSliceInjector struct {
}

func (s *customizedInterfaceSliceInjector) RuleName() string {
	return "Customized_Interface_Slice"
}

func (s *customizedInterfaceSliceInjector) Filter(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Slice && d.Type.Elem().Kind() == reflect.Interface //ruleSliceInterface
}

func (s *customizedInterfaceSliceInjector) Inject(r registry.Registry, d *meta.Node) error {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.InterfaceType(d.Type.Elem()))
	if len(metas) < 1 {
		return fmt.Errorf("none instance found for customized tag [%s] implement the interface: %s", d.Tag, d.Type)
	}
	d.Inject(metas...)
	return nil
}
