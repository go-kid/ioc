package syslog

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogLevelFiltering_New(t *testing.T) {
	// New(LvWarn): Info/Debug should not produce output, Warn/Error should
	// The default logger uses its own log.Logger, so we verify methods run without panic
	l := New(LvWarn)
	l.Info("should not panic")
	l.Debug("should not panic")
	l.Infof("format %s", "ok")
	l.Debugf("format %s", "ok")
	// Warn and Error should execute (we can't easily capture default logger output)
	l.Warn("warn message")
	l.Error("error message")
}

func TestLogLevelFiltering_NewSlogAdapter(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := NewSlogAdapter(handler).Level(LvWarn)

	buf.Reset()
	l.Info("info message")
	assert.Empty(t, buf.String(), "Info should not produce output when level is Warn")

	l.Debug("debug message")
	assert.Empty(t, buf.String(), "Debug should not produce output when level is Warn")

	l.Warn("warn message")
	assert.Contains(t, buf.String(), "warn message", "Warn should produce output")

	buf.Reset()
	l.Error("error message")
	assert.Contains(t, buf.String(), "error message", "Error should produce output")
}

func TestPrefCaching(t *testing.T) {
	SetLogger(New(LvInfo))
	// Use unique key to avoid cache pollution from other tests
	key := "pref_cache_test_unique"
	l1 := Pref(key)
	l2 := Pref(key)
	assert.Same(t, l1, l2, "Pref should return same instance for same key (caching)")
}

func TestNewLvFromString(t *testing.T) {
	tests := []struct {
		s    string
		want Lv
	}{
		{"info", LvInfo},
		{"debug", LvDebug},
		{"trace", LvTrace},
		{"warn", LvWarn},
		{"error", LvError},
		{"panic", LvPanic},
		{"fatal", LvFatal},
		{"unknown", Lv(0)},
		{"", Lv(0)},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := NewLvFromString(tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLv_String(t *testing.T) {
	assert.Equal(t, "info", LvInfo.String())
	assert.Equal(t, "debug", LvDebug.String())
	assert.Equal(t, "error", LvError.String())
	assert.Equal(t, "warn", LvWarn.String())
	assert.Equal(t, "trace", LvTrace.String())
	assert.Equal(t, "panic", LvPanic.String())
	assert.Equal(t, "fatal", LvFatal.String())
}

func TestNewSlogAdapter_ImplementsLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := NewSlogAdapter(handler)
	// Verify it implements Logger interface by calling methods
	var _ Logger = l
	l.Info("test")
	assert.Contains(t, buf.String(), "test")
}
