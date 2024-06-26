package func_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
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

type ValuedInitCompB struct {
	InterfaceImplComponent
}

func (b *ValuedInitCompB) Comp() string {
	return "B"
}

type InterfaceImplComponent struct {
	Name string
}

func (b *InterfaceImplComponent) SimpleInterface() {}

func TestCustomizedTagInject(t *testing.T) {
	t.Run("No_Value_Pointer", func(t *testing.T) {
		var tApp = &struct {
			T *NoValueComp `func:"Comp"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				&NoValueComp{"NoValueComp"},
				&ValuedComp{"ValuedComp"},
				tApp,
			))
		assert.Equal(t, "NoValueComp", tApp.T.Name)
	})
	t.Run("Valued_Pointer", func(t *testing.T) {
		var tApp = &struct {
			T *ValuedComp `func:"Comp,returns=A"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				&NoValueComp{"NoValueComp"},
				&ValuedComp{"ValuedComp"},
				tApp,
			))
		assert.Equal(t, "ValuedComp", tApp.T.Name)
	})
	t.Run("No_Value_Interface", func(t *testing.T) {
		var tApp = &struct {
			T Interface `func:"Comp"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
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
			T Interface `func:"Comp,returns=A"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
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
			T Interface `func:"Comp,returns=*"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				tApp,
			))
		assert.Equal(t, "ValuedInitComp", tApp.T.(*ValuedInitComp).Name)
	})
	t.Run("No_Value_Slice", func(t *testing.T) {
		var tApp = &struct {
			T []Interface `func:"Comp"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				&NoValueInitComp{InterfaceImplComponent{"NoValueInitComp"}},
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				tApp,
			))
		assert.Equal(t, 1, len(tApp.T))
		assert.Equal(t, "NoValueInitComp", tApp.T[0].(*NoValueInitComp).Name)
	})
	t.Run("Valued_Slice", func(t *testing.T) {
		var tApp = &struct {
			T []Interface `func:"Comp,returns=A"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				&NoValueInitComp{InterfaceImplComponent{"NoValueInitComp"}},
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				tApp,
			))
		assert.Equal(t, 1, len(tApp.T))
		assert.Equal(t, "ValuedInitComp", tApp.T[0].(*ValuedInitComp).Name)
	})
	t.Run("Ignore_Valued_Slice", func(t *testing.T) {
		var tApp = &struct {
			T []Interface `func:"Comp,returns=*"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				&NoValueInitComp{InterfaceImplComponent{"NoValueInitComp"}},
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				tApp,
			))
		assert.Equal(t, 2, len(tApp.T))
	})
	t.Run("Multi_Valued_slice", func(t *testing.T) {
		var tApp = &struct {
			T []Interface `func:"Comp,returns=A B"`
		}{}
		ioc.RunTest(t,
			app.SetComponents(
				&NoValueInitComp{InterfaceImplComponent{"NoValueInitComp"}},
				&ValuedInitComp{InterfaceImplComponent{"ValuedInitComp"}},
				&ValuedInitCompB{InterfaceImplComponent{"ValuedInitCompB"}},
				tApp,
			))
		assert.Equal(t, 2, len(tApp.T))
	})
}
