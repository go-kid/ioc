package ioc

import (
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/defination"
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

	//sc := scanner.New("Comp")
	RunTest(t, app.SetScanTags("Comp"), app.SetComponents(m, a, b))
}
