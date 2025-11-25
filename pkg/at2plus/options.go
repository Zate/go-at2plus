package at2plus

import (
	"errors"
	"log/slog"
	"time"
)

// ClientOption configures a Client.
type ClientOption func(*clientConfig) error

// clientConfig holds the configuration for a Client.
type clientConfig struct {
	port           int
	connectTimeout time.Duration
	requestTimeout time.Duration
	logger         *slog.Logger
}

// defaultConfig returns the default client configuration.
func defaultConfig() *clientConfig {
	return &clientConfig{
		port:           9200,
		connectTimeout: 5 * time.Second,
		requestTimeout: 2 * time.Second,
		logger:         nil,
	}
}

// WithPort sets the TCP port to connect to.
// Default is 9200.
func WithPort(port int) ClientOption {
	return func(c *clientConfig) error {
		if port < 1 || port > 65535 {
			return errors.New("port must be between 1 and 65535")
		}
		c.port = port
		return nil
	}
}

// WithConnectTimeout sets the timeout for establishing a connection.
// Default is 5 seconds.
func WithConnectTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) error {
		if d <= 0 {
			return errors.New("connect timeout must be positive")
		}
		c.connectTimeout = d
		return nil
	}
}

// WithRequestTimeout sets the timeout for waiting for a response.
// Default is 2 seconds.
func WithRequestTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) error {
		if d <= 0 {
			return errors.New("request timeout must be positive")
		}
		c.requestTimeout = d
		return nil
	}
}

// WithLogger sets a structured logger for debug and error logging.
// By default, no logging is performed.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *clientConfig) error {
		c.logger = logger
		return nil
	}
}
