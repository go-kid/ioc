package modified_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/factory"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"testing"
)

type MyInjector struct {
}

func (m *MyInjector) Priority() int {
	return 0
}

func (m *MyInjector) RuleName() string {
	return "My_Injector"
}

func (m *MyInjector) Condition(d *component_definition.Node) bool {
	return d.Tag == "mul"
}

type scanCompPostProcessor struct {
}

func (s *scanCompPostProcessor) PostProcessDefinitionRegistry(registry factory.DefinitionRegistry, component any, componentName string) error {
	meta := registry.GetMetaOrRegister(componentName, func() *component_definition.Meta {
		return component_definition.NewMeta(component)
	})
	for _, field := range meta.Fields {
		if tagVal, ok := field.StructField.Tag.Lookup("mul"); ok {
			meta.SetNodes(component_definition.NewNode(field, component_definition.NodeTypeComponent, "mul", tagVal))
		}
	}
	return nil
}

func (m *MyInjector) Candidates(_ factory.DefinitionRegistry, d *component_definition.Node) ([]*component_definition.Meta, error) {
	n, err := strconv.ParseInt(d.TagVal, 10, 64)
	if err != nil {
		return nil, err
	}
	d.Value.Set(reflect.ValueOf(func(i int64) int64 {
		return n * i
	}))
	d.SetArg(component_definition.ArgRequired, []string{"false"})
	return nil, nil
}

func TestModifiedInjector(t *testing.T) {
	var tApp = &struct {
		Mul func(i int64) int64 `mul:"2"`
	}{}
	ioc.RunTest(t,
		app.AddInjectionRules(new(MyInjector)),
		app.SetComponents(tApp, &scanCompPostProcessor{}))
	mul := tApp.Mul(2)
	assert.Equal(t, int64(4), mul)
}
