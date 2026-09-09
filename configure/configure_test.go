package configure

import (
	"errors"
	"testing"

	binderpkg "github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type loaderFunc func() ([]byte, error)

func (f loaderFunc) LoadConfig() ([]byte, error) { return f() }

type recordingLoader struct {
	name  string
	order int
	calls *[]string
}

func (l recordingLoader) LoadConfig() ([]byte, error) {
	*l.calls = append(*l.calls, l.name)
	return nil, nil
}

func (l recordingLoader) Order() int { return l.order }

type priorityRecordingLoader struct{ recordingLoader }

func (priorityRecordingLoader) Priority() {}

type errorBinder struct{ err error }

func (b *errorBinder) SetConfig([]byte) error { return b.err }
func (*errorBinder) Get(string) any           { return nil }
func (*errorBinder) Set(string, any)          {}

func TestConfigureInitializeWithoutLoaders(t *testing.T) {
	c := NewConfigure()
	require.NoError(t, c.Initialize())
}

func TestConfigureMergesLoaders(t *testing.T) {
	c := NewConfigure()
	c.SetBinder(binderpkg.NewViperBinder("yaml"))
	c.SetLoaders(
		loader.NewRawLoader([]byte("server:\n  host: localhost\n")),
		loader.NewRawLoader([]byte("server:\n  port: 8080\n")),
	)

	require.NoError(t, c.Initialize())
	assert.Equal(t, "localhost", c.Get("server.host"))
	assert.Equal(t, 8080, c.Get("server.port"))

	c.Set("server.host", "127.0.0.1")
	assert.Equal(t, "127.0.0.1", c.Get("server.host"))
}

func TestConfigureOrdersPriorityAndOrderedLoaders(t *testing.T) {
	var calls []string
	c := NewConfigure()
	c.SetBinder(binderpkg.NewViperBinder("yaml"))
	c.SetLoaders(
		recordingLoader{name: "ordinary-late", order: 20, calls: &calls},
		priorityRecordingLoader{recordingLoader{name: "priority", order: 99, calls: &calls}},
		recordingLoader{name: "ordinary-early", order: 10, calls: &calls},
	)

	require.NoError(t, c.Initialize())
	assert.Equal(t, []string{"priority", "ordinary-early", "ordinary-late"}, calls)
}

func TestConfigureWrapsLoaderAndBinderErrors(t *testing.T) {
	t.Run("loader", func(t *testing.T) {
		loadErr := errors.New("load failed")
		c := NewConfigure()
		c.SetLoaders(loaderFunc(func() ([]byte, error) { return nil, loadErr }))

		err := c.Initialize()
		assert.ErrorIs(t, err, loadErr)
		assert.ErrorContains(t, err, "loader")
	})

	t.Run("binder", func(t *testing.T) {
		bindErr := errors.New("bind failed")
		c := NewConfigure()
		c.SetBinder(&errorBinder{err: bindErr})
		c.SetLoaders(loader.NewRawLoader([]byte("key: value")))

		err := c.Initialize()
		assert.ErrorIs(t, err, bindErr)
		assert.ErrorContains(t, err, "raw configuration")
	})
}
