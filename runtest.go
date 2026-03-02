package ioc

import (
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

func RunTest(t *testing.T, ops ...app.SettingOption) *app.App {
	testing.Init()
	a := app.NewApp()
	err := a.Run(ops...)
	if t != nil {
		assert.NoError(t, err)
	}
	return a
}

func RunErrorTest(t *testing.T, ops ...app.SettingOption) *app.App {
	testing.Init()
	a := app.NewApp()
	err := a.Run(ops...)
	if t != nil {
		assert.Error(t, err)
	}
	return a
}
