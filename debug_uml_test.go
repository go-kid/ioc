package ioc

import (
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDebugUml(t *testing.T) {
	var (
		m = &customized{}
		a = &compA{}
		b = &compB{}
		c = &compC{}
	)

	//sc := scanner.New("Comp")
	_, err := RunDebug(DebugSetting{
		DisablePackageView:      false,
		DisableConfig:           false,
		DisableConfigDetail:     false,
		DisableDependency:       false,
		DisableDependencyDetail: false,
		DisableUselessClass:     false,
		PreciseArrow:            true,
		Writer:                  nil,
	}, app.SetScanTags("Comp"), app.SetComponents(m, a, b, c))
	assert.NoError(t, err)
}