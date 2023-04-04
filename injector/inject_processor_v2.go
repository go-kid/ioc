package injector

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/meta"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/util/list"
	"log"
	"reflect"
)

type injectProcessor interface {
	Filter(d *meta.Dependency) bool
	Inject(r *registry.Registry, d *meta.Dependency) error
}

var injectors = []injectProcessor{
	new(specifyInjector),
	new(unSpecifyPtrInjector),
	new(unSpecifyInterfaceInjector),
	new(unSpecifyInterfaceSliceInjector),
}

func DependencyInject(r *registry.Registry, m *meta.Meta) error {
	for _, dependency := range m.Dependencies {
		err := injectDependency(r, m.ID(), dependency)
		if err != nil {
			return err
		}
	}
	return nil
}

const diErrOutput = "DI report error by processor: %d\n" +
	"caused instance: %s\n" +
	"caused field: %s\n" +
	"caused by: %v\n"

func injectDependency(r *registry.Registry, metaID string, d *meta.Dependency) error {
	i, find := list.NewList(injectors).FindBy(func(i int) bool {
		return injectors[i].Filter(d)
	})
	if !find {
		return fmt.Errorf(diErrOutput, 0, metaID, d.Name(), "injection condition not found")
	}
	defer func() {
		if err := recover(); err != nil {
			log.Printf(diErrOutput, i+1, metaID, d.Name(), err)
		}
	}()
	err := injectors[i].Inject(r, d)
	if err != nil {
		return fmt.Errorf(diErrOutput, i+1, metaID, d.Name(), err)
	}
	return nil
}

/*
- Inject_Type: inject by name
- Inject_Rule:
- field is exported
- field is pointer or interface
- field has injectTag tag, and is not empty
*/
type specifyInjector struct{}

func (b *specifyInjector) Filter(d *meta.Dependency) bool {
	return d.SpecifyName != "" && //ruleTagNotEmpty
		(d.Type.Kind() == reflect.Ptr || d.Type.Kind() == reflect.Interface)
}

func (b *specifyInjector) Inject(r *registry.Registry, d *meta.Dependency) error {
	dm := r.GetComponentByName(d.SpecifyName)
	if dm == nil {
		return fmt.Errorf("no instance found for specify name: %s", d.SpecifyName)
	}
	d.Value.Set(dm.Value)
	return nil
}

/*
- Inject_Type: inject by type
- Inject_Rule:
- field is exported
- field is pointer
- field has injectTag tag, and is empty
*/
type unSpecifyPtrInjector struct{}

func (b *unSpecifyPtrInjector) Filter(d *meta.Dependency) bool {
	return d.SpecifyName == "" && //ruleEmptyTag
		d.Type.Kind() == reflect.Ptr //rulePointer
}

func (b *unSpecifyPtrInjector) Inject(r *registry.Registry, d *meta.Dependency) error {
	dm := r.GetComponentByName(d.Name())
	if dm == nil {
		return fmt.Errorf("no instance found for pointer type %s", d.Name())
	}
	d.Value.Set(dm.Value)
	return nil
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

func (i *unSpecifyInterfaceInjector) Filter(d *meta.Dependency) bool {
	return d.SpecifyName == "" && //ruleEmptyTag
		d.Type.Kind() == reflect.Interface //ruleInterface
}

func (i *unSpecifyInterfaceInjector) Inject(r *registry.Registry, d *meta.Dependency) error {
	metas := r.GetBeansByInterfaceType(d.Type)
	if len(metas) == 0 {
		return fmt.Errorf("no instance found implement interface: %s", d.Type.String())
	}
	var dm *meta.Meta
	//find unnamed instances first
	for _, m := range metas {
		if _, ok := m.Raw.(defination.NamingComponent); !ok {
			dm = m
			break
		}
	}
	if dm == nil {
		dm = metas[0]
	}
	d.Value.Set(dm.Value)
	return nil
}

/*
- Inject_Type: inject by interface slice
- Inject_Rule:
- field is exported
- field is interface slice
- field has injectTag tag, and is empty
*/
type unSpecifyInterfaceSliceInjector struct{}

func (s *unSpecifyInterfaceSliceInjector) Filter(d *meta.Dependency) bool {
	return d.SpecifyName == "" && //ruleEmptyTag
		d.Type.Kind() == reflect.Slice && d.Type.Elem().Kind() == reflect.Interface //ruleSliceInterface
}

func (s *unSpecifyInterfaceSliceInjector) Inject(r *registry.Registry, d *meta.Dependency) error {
	metas := r.GetBeansByInterfaceType(d.Type.Elem())
	if len(metas) == 0 {
		return nil
	}
	var values []reflect.Value
	var raws []interface{}
	for _, m := range metas {
		values = append(values, m.Value)
		raws = append(raws, m.Raw)
	}
	d.Value.Set(reflect.Append(d.Value, values...))
	return nil
}
