package ioc

import (
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/defination"
	"github.com/go-kid/ioc/scanner"
	"github.com/stretchr/testify/assert"
	"testing"
)

type customized struct {
	CompA   *compA                         `Comp:""`
	CompB   *compB                         `Comp:"compB"`
	CompB2  defination.InitializeComponent `Comp:"compB"`
	Comps   []any                          `Comp:"-"`
	Config  *config
	Config2 string `prop:"path"`
}

type config struct {
	C1  string
	C2  int
	Cfg *config
}

func (c *config) Prefix() string {
	return "path"
}

type compA struct {
	baseProcessor
}

func (a *compA) Comp() {}

type compB struct {
	baseProcessor
}

func (b *compB) Comp() string {
	return "compB"
}

type baseProcessor struct {
}

func (b *baseProcessor) Init() error {
	return nil
}

func TestCustomizedTagInject(t *testing.T) {
	var (
		m = &customized{}
		a = &compA{}
		b = &compB{}
	)

	sc := scanner.New("Comp")
	//meta := sc.ScanComponent(m)
	//RunTest(t, app.SetComponents(m, a, b))
	//fmt.Println(meta)
	_, err := RunDebug(DebugSetting{
		DisableConfig:           false,
		DisableConfigDetail:     true,
		DisableDependency:       false,
		DisableDependencyDetail: true,
		DisableUselessClass:     false,
		PreciseArrow:            false,
		Writer:                  nil,
	}, app.SetScanner(sc), app.SetComponents(m, a, b))
	assert.NoError(t, err)
}
