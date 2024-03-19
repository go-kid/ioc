package injector

import (
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"reflect"
)

const (
	PriorityCustomizedPtrInjector = iota + 2000
	PriorityCustomizedInterfaceInjector
	PriorityCustomizedInterfaceSliceInjector
)

type customizedPtrInjector struct {
}

func (c *customizedPtrInjector) Priority() int {
	return PriorityCustomizedPtrInjector
}

func (c *customizedPtrInjector) RuleName() string {
	return "Customized_Pointer"
}

func (c *customizedPtrInjector) Filter(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Ptr
}

func (c *customizedPtrInjector) Inject(r registry.Registry, d *meta.Node) error {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.Type(d.Type))
	return d.Inject(metas)
}

type customizedInterfaceInjector struct {
}

func (c *customizedInterfaceInjector) Priority() int {
	return PriorityCustomizedInterfaceInjector
}

func (c *customizedInterfaceInjector) RuleName() string {
	return "Customized_Interface"
}

func (c *customizedInterfaceInjector) Filter(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Interface
}

func (c *customizedInterfaceInjector) Inject(r registry.Registry, d *meta.Node) error {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.InterfaceType(d.Type))
	return d.Inject(metas)
}

type customizedInterfaceSliceInjector struct {
}

func (s *customizedInterfaceSliceInjector) Priority() int {
	return PriorityCustomizedInterfaceSliceInjector
}

func (s *customizedInterfaceSliceInjector) RuleName() string {
	return "Customized_Interface_Slice"
}

func (s *customizedInterfaceSliceInjector) Filter(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Slice && d.Type.Elem().Kind() == reflect.Interface //ruleSliceInterface
}

func (s *customizedInterfaceSliceInjector) Inject(r registry.Registry, d *meta.Node) error {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.InterfaceType(d.Type.Elem()))
	return d.Inject(metas)
}
