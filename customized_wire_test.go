package ioc

import (
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/scanner"
	"github.com/stretchr/testify/assert"
	"testing"
)

type customized struct {
	CompA *compA `Comp:""`
	CompB *compB `Comp:""`
}

type compA struct {
}

func (a *compA) Comp() {}

type compB struct {
}

func (b *compB) Comp() string {
	return "compB"
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
	_, err := RunDebug(app.SetComponents(m, a, b), app.SetScanner(sc))
	assert.NoError(t, err)
}
