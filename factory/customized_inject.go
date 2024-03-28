package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/registry"
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

func (c *customizedPtrInjector) Condition(d *component_definition.Node) bool {
	return d.Type.Kind() == reflect.Ptr
}

func (c *customizedPtrInjector) Candidates(r registry.Registry, d *component_definition.Node) ([]*component_definition.Meta, error) {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.Type(d.Type))
	return metas, nil
}

type customizedInterfaceInjector struct {
}

func (c *customizedInterfaceInjector) Priority() int {
	return PriorityCustomizedInterfaceInjector
}

func (c *customizedInterfaceInjector) RuleName() string {
	return "Customized_Interface"
}

func (c *customizedInterfaceInjector) Condition(d *component_definition.Node) bool {
	return d.Type.Kind() == reflect.Interface
}

func (c *customizedInterfaceInjector) Candidates(r registry.Registry, d *component_definition.Node) ([]*component_definition.Meta, error) {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.InterfaceType(d.Type))
	return metas, nil
}

type customizedInterfaceSliceInjector struct {
}

func (s *customizedInterfaceSliceInjector) Priority() int {
	return PriorityCustomizedInterfaceSliceInjector
}

func (s *customizedInterfaceSliceInjector) RuleName() string {
	return "Customized_Interface_Slice"
}

func (s *customizedInterfaceSliceInjector) Condition(d *component_definition.Node) bool {
	return d.Type.Kind() == reflect.Slice && d.Type.Elem().Kind() == reflect.Interface //ruleSliceInterface
}

func (s *customizedInterfaceSliceInjector) Candidates(r registry.Registry, d *component_definition.Node) ([]*component_definition.Meta, error) {
	metas := r.GetComponents(registry.FuncNameAndResult(d.Tag, d.TagVal), registry.InterfaceType(d.Type.Elem()))
	return metas, nil
}
