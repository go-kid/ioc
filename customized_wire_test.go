package ioc

import (
	"fmt"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/scanner"
	"testing"
)

type customized struct {
	CompA *compA `comp:""`
	CompB *compB `comp:""`
}

type compA struct {
}

type compB struct {
}

func TestCustomizedTagInject(t *testing.T) {
	var (
		m = &customized{}
		a = &compA{}
		b = &compB{}
	)
	meta := scanner.New("comp").
		ScanComponent(m)
	RunTest(t, app.SetComponents(m, a, b))
	fmt.Println(meta)
}
