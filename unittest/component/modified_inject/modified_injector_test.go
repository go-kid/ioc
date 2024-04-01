package modified_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"testing"
)

type MyInjector struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
}

func (m *MyInjector) Order() int {
	return 100
}

func (m *MyInjector) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}

func (m *MyInjector) PostProcessProperties(properties []*component_definition.Node, component any, componentName string) ([]*component_definition.Node, error) {
	for _, d := range properties {
		if d.Tag != "mul" {
			continue
		}
		n, err := strconv.ParseInt(d.TagVal, 10, 64)
		if err != nil {
			return nil, err
		}
		d.Value.Set(reflect.ValueOf(func(i int64) int64 {
			return n * i
		}))
		d.SetArg(component_definition.ArgRequired, []string{"false"})
	}
	return nil, nil
}

type scanCompPostProcessor struct {
	processors.DefaultTagScanDefinitionRegistryPostProcessor
}

func TestModifiedInjector(t *testing.T) {
	var tApp = &struct {
		Mul func(i int64) int64 `mul:"2"`
	}{}
	ioc.RunTest(t,
		app.SetComponents(tApp,
			&MyInjector{},
			&scanCompPostProcessor{
				processors.DefaultTagScanDefinitionRegistryPostProcessor{
					NodeType: "function",
					Tag:      "mul",
				},
			}))
	mul := tApp.Mul(2)
	assert.Equal(t, int64(4), mul)
}
