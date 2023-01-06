package tinyhttp

import (
	"crypto/tls"
	"time"
)

// ServerConfig holds a configuration for NewServer.
type ServerConfig struct {
	address string

	// Network is a network type for the listener (default: "tcp").
	Network string

	// GinMode defines working mode of the Gin library (default: Release).
	GinMode string

	// SecurityHeaders defines whether to include HTTP security headers to all responses or not (default: true).
	SecurityHeaders bool

	// MethodNotAllowed defines whether the library handles method not allowed errors (default: false).
	MethodNotAllowed bool

	// ShutdownTimeout defines a maximal timeout of HTTP server shutdown (default: 5s).
	ShutdownTimeout time.Duration

	// TLSCert is a path to TLS certificate to use. When specified with TLSKey - enables TLS mode.
	TLSCert string

	// TLSKey is a path to TLS key to use. When specified with TLSCert - enables TLS mode.
	TLSKey string

	// TLSConfig is an optional TLS configuration to pass when using TLS mode.
	TLSConfig *tls.Config

	// ReadTimeout is a timeout used when creating underlying http server (see http.Server) (default: 5s).
	ReadTimeout time.Duration

	// ReadHeaderTimeout is a timeout used when creating underlying http server (see http.Server).
	ReadHeaderTimeout time.Duration

	// WriteTimeout is a timeout used when creating underlying http server (see http.Server) (default: 10s).
	WriteTimeout time.Duration

	// IdleTimeout is a timeout used when creating underlying http server (see http.Server) (default 2m).
	IdleTimeout time.Duration

	// MaxHeaderBytes is a value used when creating underlying http server (see http.Server).
	MaxHeaderBytes int

	// TrustedProxies is a list of CIDR address ranges that can be trusted when handling RemoteIP header.
	// (default: "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8", "fc00::/7", "::1/128")
	TrustedProxies []string

	// RemoteIPHeaders is a list of headers that overwrite the value of client's remote address.
	// (default: "X-Forwarded-For")
	RemoteIPHeaders []string
}

// ServerOpt is an option to be specified to NewServer.
type ServerOpt = func(*ServerConfig)

// Network is a network type for the listener.
func Network(network string) ServerOpt {
	return func(config *ServerConfig) {
		config.Network = network
	}
}

// GinMode defines working mode of the Gin library.
func GinMode(mode string) ServerOpt {
	return func(config *ServerConfig) {
		config.GinMode = mode
	}
}

// SecurityHeaders defines whether to include HTTP security headers to all responses or not.
func SecurityHeaders(securityHeaders bool) ServerOpt {
	return func(config *ServerConfig) {
		config.SecurityHeaders = securityHeaders
	}
}

// MethodNotAllowed defines whether the library handles method not allowed errors.
func MethodNotAllowed(methodNotAllowed bool) ServerOpt {
	return func(config *ServerConfig) {
		config.MethodNotAllowed = methodNotAllowed
	}
}

// ShutdownTimeout defines a maximal timeout of HTTP server shutdown.
func ShutdownTimeout(timeout time.Duration) ServerOpt {
	return func(config *ServerConfig) {
		config.ShutdownTimeout = timeout
	}
}

// TLS enables TLS mode if both cert and key point to valid TLS credentials. tlsConfig is optional.
func TLS(cert, key string, tlsConfig ...*tls.Config) ServerOpt {
	return func(config *ServerConfig) {
		config.TLSCert = cert
		config.TLSKey = key

		if len(tlsConfig) > 0 {
			config.TLSConfig = tlsConfig[0]
		}
	}
}

// ReadTimeout is a timeout used when creating underlying http server (see http.Server).
func ReadTimeout(timeout time.Duration) ServerOpt {
	return func(config *ServerConfig) {
		config.ReadTimeout = timeout
	}
}

// ReadHeaderTimeout is a timeout used when creating underlying http server (see http.Server).
func ReadHeaderTimeout(timeout time.Duration) ServerOpt {
	return func(config *ServerConfig) {
		config.ReadHeaderTimeout = timeout
	}
}

// WriteTimeout is a timeout used when creating underlying http server (see http.Server).
func WriteTimeout(timeout time.Duration) ServerOpt {
	return func(config *ServerConfig) {
		config.WriteTimeout = timeout
	}
}

// IdleTimeout is a timeout used when creating underlying http server (see http.Server).
func IdleTimeout(timeout time.Duration) ServerOpt {
	return func(config *ServerConfig) {
		config.IdleTimeout = timeout
	}
}

// MaxHeaderBytes is a value used when creating underlying http server (see http.Server).
func MaxHeaderBytes(maxHeaderBytes int) ServerOpt {
	return func(config *ServerConfig) {
		config.MaxHeaderBytes = maxHeaderBytes
	}
}

// TrustedProxies is a list of CIDR address ranges that can be trusted when handling RemoteIP header.
func TrustedProxies(trustedProxies []string) ServerOpt {
	return func(config *ServerConfig) {
		config.TrustedProxies = trustedProxies
	}
}

// RemoteIPHeaders is a list of headers that overwrite the value of client's remote address.
func RemoteIPHeaders(headers []string) ServerOpt {
	return func(config *ServerConfig) {
		config.RemoteIPHeaders = headers
	}
}
