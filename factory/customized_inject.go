package factory

import (
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

func (c *customizedPtrInjector) Condition(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Ptr
}

func (c *customizedPtrInjector) Candidates(r BuildContainer, d *meta.Node) ([]*meta.Meta, error) {
	metas := r.GetMetas(FuncNameAndResult(d.Tag, d.TagVal), Type(d.Type))
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

func (c *customizedInterfaceInjector) Condition(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Interface
}

func (c *customizedInterfaceInjector) Candidates(r BuildContainer, d *meta.Node) ([]*meta.Meta, error) {
	metas := r.GetMetas(FuncNameAndResult(d.Tag, d.TagVal), InterfaceType(d.Type))
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

func (s *customizedInterfaceSliceInjector) Condition(d *meta.Node) bool {
	return d.Type.Kind() == reflect.Slice && d.Type.Elem().Kind() == reflect.Interface //ruleSliceInterface
}

func (s *customizedInterfaceSliceInjector) Candidates(r BuildContainer, d *meta.Node) ([]*meta.Meta, error) {
	metas := r.GetMetas(FuncNameAndResult(d.Tag, d.TagVal), InterfaceType(d.Type.Elem()))
	return metas, nil
}
