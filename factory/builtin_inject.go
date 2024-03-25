package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"reflect"
)

const (
	PrioritySpecifyInjector = iota + 1000
	PriorityUnSpecifyPtrInjector
	PriorityUnSpecifyPtrSliceInjector
	PriorityUnSpecifyInterfaceInjector
	PriorityUnSpecifyInterfaceSliceInjector
)

/*
- Inject_Type: inject by name
- Inject_Rule:
- field is exported
- field is pointer or interface
- field has injectTag tag, and is not empty
*/
type specifyInjector struct{}

func (b *specifyInjector) Priority() int {
	return PrioritySpecifyInjector
}

func (b *specifyInjector) RuleName() string {
	return "Any_Type_With_Specifying_Name"
}

func (b *specifyInjector) Condition(d *component_definition.Node) bool {
	return d.Tag == definition.InjectTag && d.TagVal != "" && //ruleTagNotEmpty
		(d.Type.Kind() == reflect.Ptr || d.Type.Kind() == reflect.Interface)
}

func (b *specifyInjector) Candidates(r DefinitionRegistry, d *component_definition.Node) ([]*component_definition.Meta, error) {
	dm := r.GetMetaByName(d.TagVal)
	return []*component_definition.Meta{dm}, nil
}

/*
- Inject_Type: inject by type
- Inject_Rule:
- field is exported
- field is pointer
- field has injectTag tag, and is empty
*/
type unSpecifyPtrInjector struct{}

func (b *unSpecifyPtrInjector) Priority() int {
	return PriorityUnSpecifyPtrInjector
}

func (b *unSpecifyPtrInjector) RuleName() string {
	return "Pointer_Without_Specifying_Name"
}

func (b *unSpecifyPtrInjector) Condition(d *component_definition.Node) bool {
	return d.Tag == definition.InjectTag && d.TagVal == "" && //ruleEmptyTag
		d.Type.Kind() == reflect.Ptr //rulePointer
}

func (b *unSpecifyPtrInjector) Candidates(r DefinitionRegistry, d *component_definition.Node) ([]*component_definition.Meta, error) {
	metas := r.GetMetas(Type(d.Type))
	return metas, nil
}

/*
- Inject_Type: inject by Ptr slice
- Inject_Rule:
- field is exported
- field is Ptr slice
- field has injectTag tag, and is empty
*/
type unSpecifyPtrSliceInjector struct{}

func (s *unSpecifyPtrSliceInjector) Priority() int {
	return PriorityUnSpecifyPtrSliceInjector
}

func (s *unSpecifyPtrSliceInjector) RuleName() string {
	return "Pointer_Slice"
}

func (s *unSpecifyPtrSliceInjector) Condition(d *component_definition.Node) bool {
	return d.Tag == definition.InjectTag && d.TagVal == "" && //ruleEmptyTag
		d.Type.Kind() == reflect.Slice && d.Type.Elem().Kind() == reflect.Pointer //ruleSlicePtr
}

func (s *unSpecifyPtrSliceInjector) Candidates(r DefinitionRegistry, d *component_definition.Node) ([]*component_definition.Meta, error) {
	metas := r.GetMetas(Type(d.Type.Elem()))
	return metas, nil
}

/*
- Inject_Type: inject by interface
- Inject_Rule:
- field is exported
- field is interface
- field has injectTag tag, and is empty
- prefer the first unnamed(not implement NamingComponent) instance
*/
type unSpecifyInterfaceInjector struct{}

func (i *unSpecifyInterfaceInjector) Priority() int {
	return PriorityUnSpecifyInterfaceInjector
}

func (i *unSpecifyInterfaceInjector) RuleName() string {
	return "Interface_Without_Specifying_Name"
}

func (i *unSpecifyInterfaceInjector) Condition(d *component_definition.Node) bool {
	return d.Tag == definition.InjectTag && d.TagVal == "" && //ruleEmptyTag
		d.Type.Kind() == reflect.Interface //ruleInterface
}

func (i *unSpecifyInterfaceInjector) Candidates(r DefinitionRegistry, d *component_definition.Node) ([]*component_definition.Meta, error) {
	metas := r.GetMetas(InterfaceType(d.Type))
	return metas, nil
}

/*
- Inject_Type: inject by interface slice
- Inject_Rule:
- field is exported
- field is interface slice
- field has injectTag tag, and is empty
*/
type unSpecifyInterfaceSliceInjector struct{}

func (s *unSpecifyInterfaceSliceInjector) Priority() int {
	return PriorityUnSpecifyInterfaceSliceInjector
}

func (s *unSpecifyInterfaceSliceInjector) RuleName() string {
	return "Interface_Slice"
}

func (s *unSpecifyInterfaceSliceInjector) Condition(d *component_definition.Node) bool {
	return d.Tag == definition.InjectTag && d.TagVal == "" && //ruleEmptyTag
		d.Type.Kind() == reflect.Slice && d.Type.Elem().Kind() == reflect.Interface //ruleSliceInterface
}

func (s *unSpecifyInterfaceSliceInjector) Candidates(r DefinitionRegistry, d *component_definition.Node) ([]*component_definition.Meta, error) {
	metas := r.GetMetas(InterfaceType(d.Type.Elem()))
	return metas, nil
}
