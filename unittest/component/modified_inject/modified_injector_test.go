package modified_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/scanner"
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

func (m *MyInjector) Candidates(_ factory.BuildContainer, d *component_definition.Node) ([]*component_definition.Meta, error) {
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
		app.AddScanPolicies(scanner.NewComponentScanPolicy("mul", nil)),
		app.AddInjectionRules(new(MyInjector)),
		app.SetComponents(tApp))
	mul := tApp.Mul(2)
	assert.Equal(t, int64(4), mul)
}
