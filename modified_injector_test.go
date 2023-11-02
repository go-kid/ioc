package ioc

import (
	"fmt"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"testing"
	"time"
)

type myInjector struct {
}

func (m *myInjector) Filter(d *meta.Node) bool {
	return d.Tag == "time"
}

func (m *myInjector) Inject(r registry.Registry, d *meta.Node) error {
	format := time.Now().Format(d.TagVal)
	d.Value.SetString(format)
	return nil
}

type tInject struct {
	Time string `time:"2006-01-02"`
}

func TestModifiedInjector(t *testing.T) {
	//sc := scanner.New("time")
	injector.AddModifyInjectors([]injector.InjectProcessor{new(myInjector)})
	ti := &tInject{}
	RunTest(t, app.SetScanTags("time"), app.SetComponents(ti))
	fmt.Println(ti.Time)
}
