package modified_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
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

func (m *MyInjector) Filter(d *meta.Node) bool {
	return d.Tag == "mul"
}

func (m *MyInjector) Inject(_ registry.Registry, d *meta.Node) error {
	n, err := strconv.ParseInt(d.TagVal, 10, 64)
	if err != nil {
		return err
	}
	d.Value.Set(reflect.ValueOf(func(i int64) int64 {
		return n * i
	}))
	return nil
}

func TestModifiedInjector(t *testing.T) {
	var tApp = &struct {
		Mul func(i int64) int64 `mul:"2"`
	}{}
	ioc.RunTest(t,
		app.SetScanTags("mul"),
		app.AddCustomizedInjectors(new(MyInjector)),
		app.SetComponents(tApp))
	mul := tApp.Mul(2)
	assert.Equal(t, int64(4), mul)
}
