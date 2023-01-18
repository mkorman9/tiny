package requests

import (
	"crypto/tls"
	"net/http"
	"time"
)

// Config holds a configuration for Client.
type Config struct {
	// Network is a network type to use (default: "tcp").
	Network string

	// Address is an optional property that overwrites network address that clients connect to when sending request.
	// By default, it's empty and the target address is extracted from the URL, and it might differ between requests.
	// Setting Address might be required if the client targets Unix socket.
	// In this case you need to set Network parameter to "unix" and Address to a path to a Unix socket
	// (for example "/run/httpd.sock").
	Address string

	// Timeout is a time after which a request (call to Client.Send()) times out (default: 10s).
	Timeout time.Duration

	// MaxRetries is a maximum number of time the request should be retried when encountering server error.
	// Requests are only retried when send operation returns network error or HTTP server errors (5xx).
	// (default: 0).
	MaxRetries int

	// MaxRedirects is a maximum number of redirects the client will perform before failing with ErrRedirect.
	// (default: 10).
	MaxRedirects int

	// RetryDelayFactor is a factor used to calculate the delay time between subsequent retries.
	// The formula is: retryNumber * RetryDelayFactor.
	// (default: 0).
	RetryDelayFactor time.Duration

	// TLSConfig is an optional TLS configuration to pass when using TLS.
	TLSConfig *tls.Config

	// CookieJar is a collection of cookies to use in all requests initiated by the client.
	CookieJar http.CookieJar
}

func mergeConfig(provided *Config) *Config {
	config := &Config{
		Network:      "tcp",
		Timeout:      10 * time.Second,
		MaxRedirects: 10,
		TLSConfig:    &tls.Config{},
	}

	if provided == nil {
		return config
	}

	if provided.Network != "" {
		config.Network = provided.Network
	}
	if provided.Address != "" {
		config.Address = provided.Address
	}
	if provided.Timeout != 0 {
		config.Timeout = provided.Timeout
	}
	if provided.MaxRetries != 0 {
		config.MaxRetries = provided.MaxRetries
	}
	if provided.MaxRedirects != 0 {
		config.MaxRedirects = provided.MaxRedirects
	}
	if provided.RetryDelayFactor != 0 {
		config.RetryDelayFactor = provided.RetryDelayFactor
	}
	if provided.TLSConfig != nil {
		config.TLSConfig = provided.TLSConfig
	}
	if provided.CookieJar != nil {
		config.CookieJar = provided.CookieJar
	}

	return config
}
