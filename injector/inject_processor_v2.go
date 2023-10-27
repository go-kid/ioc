package injector

import (
	"fmt"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/util/list"
	"log"
	"reflect"
)

type Injector interface {
	GetByName(name string) (reflect.Value, bool)
	GetOneByInterfaceType(typ reflect.Type) (reflect.Value, bool)
	GetsByInterfaceType(typ reflect.Type) []reflect.Value
}

type injectProcessor interface {
	Filter(d *defination.Node) bool
	Inject(r Injector, d *defination.Node) error
}

var injectors = []injectProcessor{
	new(specifyInjector),
	new(unSpecifyPtrInjector),
	new(unSpecifyInterfaceInjector),
	new(unSpecifyInterfaceSliceInjector),
}

func DependencyInject(r Injector, id string, dependencies []*defination.Node) error {
	for _, dependency := range dependencies {
		err := injectDependency(r, id, dependency)
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

func injectDependency(r Injector, metaID string, d *defination.Node) error {
	i, find := list.NewList(injectors).FindBy(func(i int) bool {
		return injectors[i].Filter(d)
	})
	if !find {
		return fmt.Errorf(diErrOutput, 0, metaID, d.Id(), "injection condition not found")
	}
	defer func() {
		if err := recover(); err != nil {
			log.Printf(diErrOutput, i+1, metaID, d.Id(), err)
		}
	}()
	err := injectors[i].Inject(r, d)
	if err != nil {
		return fmt.Errorf(diErrOutput, i+1, metaID, d.Id(), err)
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

func (b *specifyInjector) Filter(d *defination.Node) bool {
	return d.Tag != "" && //ruleTagNotEmpty
		(d.Type.Kind() == reflect.Ptr || d.Type.Kind() == reflect.Interface)
}

func (b *specifyInjector) Inject(r Injector, d *defination.Node) error {
	dm, ok := r.GetByName(d.Tag)
	if !ok {
		return fmt.Errorf("no instance found for specify name: %s", d.Tag)
	}
	d.Value.Set(dm)
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

func (b *unSpecifyPtrInjector) Filter(d *defination.Node) bool {
	return d.Tag == "" && //ruleEmptyTag
		d.Type.Kind() == reflect.Ptr //rulePointer
}

func (b *unSpecifyPtrInjector) Inject(r Injector, d *defination.Node) error {
	dm, ok := r.GetByName(d.Id())
	if !ok {
		return fmt.Errorf("no instance found for pointer type %s", d.Id())
	}
	d.Value.Set(dm)
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

func (i *unSpecifyInterfaceInjector) Filter(d *defination.Node) bool {
	return d.Tag == "" && //ruleEmptyTag
		d.Type.Kind() == reflect.Interface //ruleInterface
}

func (i *unSpecifyInterfaceInjector) Inject(r Injector, d *defination.Node) error {
	dm, ok := r.GetOneByInterfaceType(d.Type)
	if !ok {
		return fmt.Errorf("no instance found implement interface: %s", d.Type.String())
	}
	d.Value.Set(dm)
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

func (s *unSpecifyInterfaceSliceInjector) Filter(d *defination.Node) bool {
	return d.Tag == "" && //ruleEmptyTag
		d.Type.Kind() == reflect.Slice && d.Type.Elem().Kind() == reflect.Interface //ruleSliceInterface
}

func (s *unSpecifyInterfaceSliceInjector) Inject(r Injector, d *defination.Node) error {
	vals := r.GetsByInterfaceType(d.Type.Elem())
	if len(vals) == 0 {
		return nil
	}
	d.Value.Set(reflect.Append(d.Value, vals...))
	return nil
}
