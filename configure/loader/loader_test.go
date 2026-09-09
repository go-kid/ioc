package loader

import (
	"path/filepath"
	"testing"

	"github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
)

func TestRawLoader(t *testing.T) {
	want := []byte("key: value")
	got, err := NewRawLoader(want).LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestArgsLoader(t *testing.T) {
	data, err := NewArgsLoader([]string{
		"command",
		"--ignored=value",
		"--app.config=server.port=8080",
		"--app.config=feature.enabled=true",
		"--app.config=service.name=ioc",
	}).LoadConfig()
	require.NoError(t, err)
	require.NotEmpty(t, data)

	b := binder.NewViperBinder("yaml")
	require.NoError(t, b.SetConfig(data))
	assert.Equal(t, 8080, b.Get("server.port"))
	assert.Equal(t, true, b.Get("feature.enabled"))
	assert.Equal(t, "ioc", b.Get("service.name"))
}

func TestArgsLoaderWithoutConfig(t *testing.T) {
	data, err := NewArgsLoader([]string{"command", "--other=value"}).LoadConfig()
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestFileLoader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("key: value"), 0o600))

	l := NewFileLoader(path)
	_, priority := any(l).(definition.Priority)
	assert.True(t, priority)
	assert.Equal(t, 0, l.Order())

	data, err := l.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, []byte("key: value"), data)
}

func TestFileLoaderMissingFile(t *testing.T) {
	_, err := NewFileLoader(filepath.Join(t.TempDir(), "missing.yaml")).LoadConfig()
	assert.ErrorContains(t, err, "read file")
}
