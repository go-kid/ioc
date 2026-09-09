package binder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViperBinderDefaultsToYAMLAndMerges(t *testing.T) {
	b := NewViperBinder("")
	require.NoError(t, b.SetConfig([]byte("server:\n  host: localhost\n")))
	require.NoError(t, b.SetConfig([]byte("server:\n  port: 8080\n")))

	assert.Equal(t, "localhost", b.Get("server.host"))
	assert.Equal(t, 8080, b.Get("server.port"))
	assert.Contains(t, b.Get(""), "server")

	b.Set("server.host", "127.0.0.1")
	assert.Equal(t, "127.0.0.1", b.Get("server.host"))
}

func TestViperBinderSupportsJSON(t *testing.T) {
	b := NewViperBinder("json")
	require.NoError(t, b.SetConfig([]byte(`{"enabled":true}`)))
	assert.Equal(t, true, b.Get("enabled"))
}

func TestViperBinderRejectsInvalidConfig(t *testing.T) {
	err := NewViperBinder("yaml").SetConfig([]byte("key: ["))
	assert.ErrorContains(t, err, "viper merge config")
}
