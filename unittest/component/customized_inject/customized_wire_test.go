package customized_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/factory"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Interface interface {
	SimpleInterface()
}

type NoValueComp struct {
	Name string
}

func (a *NoValueComp) Comp() {}

type ValuedComp struct {
	Name string
}

func (b *ValuedComp) Comp() string {
	return "A"
}

type NoValueInitComp struct {
	InterfaceImplComponent
}

func (a *NoValueInitComp) Comp() {}

type ValuedInitComp struct {
	InterfaceImplComponent
}

func (b *ValuedInitComp) Comp() string {
	return "A"
}

type InterfaceImplComponent struct {
	Name string
}

func (b *InterfaceImplComponent) SimpleInterface() {}

type scanCompPostProcessor struct {
}

func (s *scanCompPostProcessor) PostProcessDefinitionRegistry(registry factory.DefinitionRegistry, component any, componentName string) error {
	meta := registry.GetMetaOrRegister(componentName, func() *component_definition.Meta {
		return component_definition.NewMeta(component)
	})
	for _, field := range meta.Fields {
		if tagVal, ok := field.StructField.Tag.Lookup("Comp"); ok {
			meta.SetNodes(component_definition.NewNode(field, component_definition.NodeTypeComponent, "Comp", tagVal))
		}
	}
	return nil
}

func TestCustomizedTagInject(t *testing.T) {
	sp := &scanCompPostProcessor{}
	t.Run("No_Value_Pointer", func(t *testing.T) {
		var tApp = &struct {
			T *NoValueComp `Comp:""`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				sp,
				&NoValueComp{"NoValueComp"},
				&ValuedComp{"ValuedComp"},
				tApp,
			))
		assert.Equal(t, "NoValueComp", tApp.T.Name)
	})
	t.Run("Valued_Pointer", func(t *testing.T) {
		var tApp = &struct {
			T *ValuedComp `Comp:"A"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				sp,
				&NoValueComp{"NoValueComp"},
				&ValuedComp{"ValuedComp"},
				tApp,
			))
		assert.Equal(t, "ValuedComp", tApp.T.Name)
	})
	t.Run("No_Value_Interface", func(t *testing.T) {
		var tApp = &struct {
			T Interface `Comp:""`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				sp,
				&NoValueComp{"NoValueComp"},
				&ValuedComp{"ValuedComp"},
				&NoValueInitComp{InterfaceImplComponent{"NoValueInitComp"}},
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				tApp,
			))
		assert.Equal(t, "NoValueInitComp", tApp.T.(*NoValueInitComp).Name)
	})
	t.Run("Valued_Interface", func(t *testing.T) {
		var tApp = &struct {
			T Interface `Comp:"A"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				sp,
				&NoValueComp{"NoValueComp"},
				&ValuedComp{"ValuedComp"},
				&NoValueInitComp{InterfaceImplComponent{"NoValueInitComp"}},
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				tApp,
			))
		assert.Equal(t, "ValuedInitComp", tApp.T.(*ValuedInitComp).Name)
	})
	t.Run("Ignore_Valued_Interface", func(t *testing.T) {
		var tApp = &struct {
			T Interface `Comp:"-"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				sp,
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				tApp,
			))
		assert.Equal(t, "ValuedInitComp", tApp.T.(*ValuedInitComp).Name)
	})
	t.Run("No_Value_Slice", func(t *testing.T) {
		var tApp = &struct {
			T []Interface `Comp:""`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				sp,
				&NoValueInitComp{InterfaceImplComponent{"NoValueInitComp"}},
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				tApp,
			))
		assert.Equal(t, 1, len(tApp.T))
		assert.Equal(t, "NoValueInitComp", tApp.T[0].(*NoValueInitComp).Name)
	})
	t.Run("Valued_Slice", func(t *testing.T) {
		var tApp = &struct {
			T []Interface `Comp:"A"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				sp,
				&NoValueInitComp{InterfaceImplComponent{"NoValueInitComp"}},
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				tApp,
			))
		assert.Equal(t, 1, len(tApp.T))
		assert.Equal(t, "ValuedInitComp", tApp.T[0].(*ValuedInitComp).Name)
	})
	t.Run("Ignore_Valued_Slice", func(t *testing.T) {
		var tApp = &struct {
			T []Interface `Comp:"-"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				sp,
				&NoValueInitComp{InterfaceImplComponent{"NoValueInitComp"}},
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				tApp,
			))
		assert.Equal(t, 2, len(tApp.T))
	})
}
