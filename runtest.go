package ioc

import (
	. "github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

func RunTest(t *testing.T, ops ...SettingOption) *App {
	s := NewApp(ops...)
	err := s.Run()
	if t != nil {
		assert.NoError(t, err)
	}
	return s
}

func RunErrorTest(t *testing.T, ops ...SettingOption) *App {
	s := NewApp(ops...)
	err := s.Run()
	if t != nil {
		assert.Error(t, err)
	}
	return s
}
