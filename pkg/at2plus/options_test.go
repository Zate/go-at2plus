package at2plus

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithPort_Valid(t *testing.T) {
	cfg := defaultConfig()

	err := WithPort(9200)(cfg)
	require.NoError(t, err)
	assert.Equal(t, 9200, cfg.port)

	err = WithPort(1)(cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, cfg.port)

	err = WithPort(65535)(cfg)
	require.NoError(t, err)
	assert.Equal(t, 65535, cfg.port)
}

func TestWithPort_Invalid(t *testing.T) {
	cfg := defaultConfig()

	err := WithPort(0)(cfg)
	assert.Error(t, err)

	err = WithPort(-1)(cfg)
	assert.Error(t, err)

	err = WithPort(65536)(cfg)
	assert.Error(t, err)
}

func TestWithConnectTimeout_Valid(t *testing.T) {
	cfg := defaultConfig()

	err := WithConnectTimeout(10 * time.Second)(cfg)
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.connectTimeout)
}

func TestWithConnectTimeout_Invalid(t *testing.T) {
	cfg := defaultConfig()

	err := WithConnectTimeout(0)(cfg)
	assert.Error(t, err)

	err = WithConnectTimeout(-1 * time.Second)(cfg)
	assert.Error(t, err)
}

func TestWithRequestTimeout_Valid(t *testing.T) {
	cfg := defaultConfig()

	err := WithRequestTimeout(5 * time.Second)(cfg)
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, cfg.requestTimeout)
}

func TestWithRequestTimeout_Invalid(t *testing.T) {
	cfg := defaultConfig()

	err := WithRequestTimeout(0)(cfg)
	assert.Error(t, err)

	err = WithRequestTimeout(-1 * time.Second)(cfg)
	assert.Error(t, err)
}

func TestWithLogger(t *testing.T) {
	cfg := defaultConfig()
	assert.Nil(t, cfg.logger)

	logger := slog.Default()
	err := WithLogger(logger)(cfg)
	require.NoError(t, err)
	assert.Equal(t, logger, cfg.logger)
}

func TestWithLogger_Nil(t *testing.T) {
	cfg := defaultConfig()
	cfg.logger = slog.Default()

	err := WithLogger(nil)(cfg)
	require.NoError(t, err)
	assert.Nil(t, cfg.logger)
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	assert.Equal(t, 9200, cfg.port)
	assert.Equal(t, 5*time.Second, cfg.connectTimeout)
	assert.Equal(t, 2*time.Second, cfg.requestTimeout)
	assert.Nil(t, cfg.logger)
}
