package ioc

import (
	. "github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

func RunTest(t *testing.T, ops ...SettingOption) *App {
	testing.Init()
	app := NewApp()
	err := app.Run(ops...)
	if t != nil {
		assert.NoError(t, err)
	}
	return app
}

func RunErrorTest(t *testing.T, ops ...SettingOption) *App {
	testing.Init()
	app := NewApp()
	err := app.Run(ops...)
	if t != nil {
		assert.Error(t, err)
	}
	return app
}
