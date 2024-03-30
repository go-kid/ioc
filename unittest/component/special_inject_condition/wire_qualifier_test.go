package special_inject_condition

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ITest interface {
	action()
}

type iTest struct {
}

func (i *iTest) action() {
}

type QualifierComponent struct {
	iTest
	name  string
	group string
}

func (c *QualifierComponent) Naming() string {
	return c.name
}

func (c *QualifierComponent) Qualifier() string {
	return c.group
}

func TestWireQualifierComponent(t *testing.T) {
	var data = []any{
		&QualifierComponent{name: "test01", group: ""},
		&QualifierComponent{name: "test02", group: ""},
		&QualifierComponent{name: "test11", group: "group1"},
		&QualifierComponent{name: "test12", group: "group1"},
		&QualifierComponent{name: "test21", group: "group2"},
		&QualifierComponent{name: "test22", group: "group2"},
	}
	t.Run("SimpleSliceInject", func(t *testing.T) {
		type T struct {
			Comps []ITest `wire:""`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(data...), app.SetComponents(tt))
		assert.Equal(t, len(data), len(tt.Comps))
	})
	t.Run("QualifierEmptySingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:",qualifier"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(data...), app.SetComponents(tt))
	})
	t.Run("QualifierEmptySliceInject", func(t *testing.T) {
		type T struct {
			Comp []ITest `wire:",qualifier"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(data...), app.SetComponents(tt))
		assert.Equal(t, 2, len(tt.Comp))
	})
	t.Run("QualifierGroup1SingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:",qualifier=group1"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(data...), app.SetComponents(tt))
	})
	t.Run("QualifierGroup1SliceInject", func(t *testing.T) {
		type T struct {
			Comp []ITest `wire:",qualifier=group1"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(data...), app.SetComponents(tt))
		assert.Equal(t, 2, len(tt.Comp))
	})
	t.Run("QualifierMultipleSingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:",qualifier=group1 group2"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(data...), app.SetComponents(tt))
	})
	t.Run("QualifierMultipleSliceInject", func(t *testing.T) {
		type T struct {
			Comp []ITest `wire:",qualifier=group1 group2"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(data...), app.SetComponents(tt))
		assert.Equal(t, 4, len(tt.Comp))
	})
	t.Run("QualifierNamingSingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:"test11,qualifier=group1"`
		}
		var tt = &T{}
		ioc.RunTest(t, app.SetComponents(data...), app.SetComponents(tt))
	})
	t.Run("QualifierUnExistNamingSingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:"test33,qualifier=group1"`
		}
		var tt = &T{}
		newApp := app.NewApp(app.SetComponents(data...), app.SetComponents(tt))
		err := newApp.Run()
		assert.Error(t, err)
	})
	t.Run("QualifierUnExistGroupSingleInject", func(t *testing.T) {
		type T struct {
			Comp ITest `wire:",qualifier=group3"`
		}
		var tt = &T{}
		newApp := app.NewApp(app.SetComponents(data...), app.SetComponents(tt))
		err := newApp.Run()
		assert.Error(t, err)
	})
	t.Run("QualifierUnExistGroupSliceInject", func(t *testing.T) {
		type T struct {
			Comp []ITest `wire:",qualifier=group3"`
		}
		var tt = &T{}
		newApp := app.NewApp(app.SetComponents(data...), app.SetComponents(tt))
		err := newApp.Run()
		assert.Error(t, err)
	})
}
