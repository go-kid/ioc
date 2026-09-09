package app

import (
	"path/filepath"
	"testing"
	"time"

	configurepkg "github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/configure/loader"
	factorypkg "github.com/go-kid/ioc/container/factory"
	"github.com/go-kid/ioc/container/support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
)

type optionComponent struct{}

func TestCoreOptions(t *testing.T) {
	a := NewApp()
	registry := support.NewRegistry()
	configuration := configurepkg.NewConfigure()
	factory := factorypkg.Default()

	Options(
		SetRegistry(registry),
		SetConfigure(configuration),
		SetFactory(factory),
		SetShutdownTimeout(3*time.Second),
		SkipRunners(),
	)(a)

	assert.Same(t, registry, a.registry)
	assert.Same(t, configuration, a.Configure)
	assert.Same(t, factory, a.Factory)
	assert.Equal(t, 3*time.Second, a.shutdownTimeout)
	assert.True(t, a.skipRunners)

	component := &optionComponent{}
	SetComponents(component)(a)
	assert.True(t, registry.ContainsSingleton("github.com/go-kid/ioc/app/optionComponent"))
}

func TestConfigurationOptions(t *testing.T) {
	t.Run("set and add loaders", func(t *testing.T) {
		a := NewApp()
		SetConfigLoader(loader.NewRawLoader([]byte("first: 1")))(a)
		AddConfigLoader(loader.NewRawLoader([]byte("second: 2")))(a)
		require.NoError(t, a.Configure.Initialize())
		assert.Equal(t, 1, a.Get("first"))
		assert.Equal(t, 2, a.Get("second"))
	})

	t.Run("set binder", func(t *testing.T) {
		a := NewApp()
		SetConfigBinder(binder.NewViperBinder("json"))(a)
		SetConfigLoader(loader.NewRawLoader([]byte(`{"enabled":true}`)))(a)
		require.NoError(t, a.Configure.Initialize())
		assert.Equal(t, true, a.Get("enabled"))
	})

	t.Run("config file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		require.NoError(t, os.WriteFile(path, []byte("fileValue: loaded"), 0o600))

		a := NewApp()
		SetConfigLoader()(a)
		SetConfig(path)(a)
		require.NoError(t, a.Configure.Initialize())
		assert.Equal(t, "loaded", a.Get("fileValue"))
	})
}

func TestGlobalSettingsQueue(t *testing.T) {
	globalOptions = nil
	t.Cleanup(func() { globalOptions = nil })

	Settings(SkipRunners(), SetShutdownTimeout(time.Second))
	require.Len(t, globalOptions, 2)

	a := NewApp()
	for _, option := range globalOptions {
		option(a)
	}
	assert.True(t, a.skipRunners)
	assert.Equal(t, time.Second, a.shutdownTimeout)
}
