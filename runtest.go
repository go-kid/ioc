package ioc

import (
	. "github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func RunTest(t *testing.T, ops ...SettingOption) *App {
	s := NewApp(append([]SettingOption{SetRegistry(registry.NewRegistry())}, ops...)...)
	err := s.Run()
	if t != nil {
		assert.NoError(t, err)
	}
	return s
}
