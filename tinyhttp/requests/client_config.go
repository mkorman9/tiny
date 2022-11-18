package requests

import (
	"crypto/tls"
	"time"
)

// ClientConfig holds a configuration for Client.
type ClientConfig struct {
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

	// RetryDelayFactor is a factor used to calculate the delay time between subsequent retries.
	// The formula is: retryNumber * RetryDelayFactor.
	// (default: 0).
	RetryDelayFactor time.Duration

	// TLSConfig is an optional TLS configuration to pass when using TLS.
	TLSConfig *tls.Config
}

// ClientOpt is an option to be passed to NewClient.
type ClientOpt = func(*ClientConfig)

// Network is a network type to use.
func Network(network string) ClientOpt {
	return func(config *ClientConfig) {
		config.Network = network
	}
}

// Address is an optional property that overwrites network address that clients connect to when sending request.
// By default, it's empty and the target address is extracted from the URL, and it might differ between requests.
// Setting Address might be required if the client targets Unix socket.
// In this case you need to set Network parameter to "unix" and Address to a path to a Unix socket
// (for example "/run/httpd.sock").
func Address(address string) ClientOpt {
	return func(config *ClientConfig) {
		config.Address = address
	}
}

// Timeout is a time after which a request (call to Client.Do()) times out.
func Timeout(timeout time.Duration) ClientOpt {
	return func(config *ClientConfig) {
		config.Timeout = timeout
	}
}

// MaxRetries is a maximum number of time the request should be retried when encountering server error.
// Requests are only retried when send operation returns network error or HTTP server errors (5xx).
func MaxRetries(maxRetries int) ClientOpt {
	return func(config *ClientConfig) {
		if maxRetries < 0 {
			maxRetries = 0
		}

		config.MaxRetries = maxRetries
	}
}

// RetryDelayFactor is a factor used to calculate the delay time between subsequent retries.
// The formula is: retryNumber * RetryDelayFactor.
func RetryDelayFactor(retryDelayFactor time.Duration) ClientOpt {
	return func(config *ClientConfig) {
		config.RetryDelayFactor = retryDelayFactor
	}
}

// TLSConfig is an optional TLS configuration to pass when using TLS.
func TLSConfig(tlsConfig *tls.Config) ClientOpt {
	return func(config *ClientConfig) {
		config.TLSConfig = tlsConfig
	}
}
