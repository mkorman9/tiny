package tinyhttp

import (
	"time"
)

// ServerConfig holds a configuration for NewServer.
type ServerConfig struct {
	address string

	// Network is a network type for the listener (default: "tcp").
	Network string

	// SecurityHeaders defines whether to include HTTP security headers to all responses or not (default: true).
	SecurityHeaders bool

	// ShutdownTimeout defines a maximal timeout of HTTP server shutdown (default: 5s).
	ShutdownTimeout time.Duration

	// TLSCert is a path to TLS certificate to use. When specified with TLSKey - enables TLS mode.
	TLSCert string

	// TLSKey is a path to TLS key to use. When specified with TLSCert - enables TLS mode.
	TLSKey string

	// ReadTimeout is a timeout used when creating underlying http server (see http.Server) (default: 5s).
	ReadTimeout time.Duration

	// WriteTimeout is a timeout used when creating underlying http server (see http.Server) (default: 10s).
	WriteTimeout time.Duration

	// IdleTimeout is a timeout used when creating underlying http server (see http.Server) (default 2m).
	IdleTimeout time.Duration

	// TrustedProxies is a list of CIDR address ranges that can be trusted when handling RemoteIP header.
	// (default: "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8", "fc00::/7", "::1/128")
	TrustedProxies []string

	// RemoteIPHeaders is a list of headers that overwrite the value of client's remote address.
	// (default: "X-Forwarded-For")
	RemoteIPHeader string
}

// ServerOpt is an option to be specified to NewServer.
type ServerOpt = func(*ServerConfig)

// Network is a network type for the listener.
func Network(network string) ServerOpt {
	return func(config *ServerConfig) {
		config.Network = network
	}
}

// SecurityHeaders defines whether to include HTTP security headers to all responses or not.
func SecurityHeaders(securityHeaders bool) ServerOpt {
	return func(config *ServerConfig) {
		config.SecurityHeaders = securityHeaders
	}
}

// ShutdownTimeout defines a maximal timeout of HTTP server shutdown.
func ShutdownTimeout(timeout time.Duration) ServerOpt {
	return func(config *ServerConfig) {
		config.ShutdownTimeout = timeout
	}
}

// TLS enables TLS mode if both cert and key point to valid TLS credentials.
func TLS(cert, key string) ServerOpt {
	return func(config *ServerConfig) {
		config.TLSCert = cert
		config.TLSKey = key
	}
}

// ReadTimeout is a timeout used when creating underlying http server (see http.Server).
func ReadTimeout(timeout time.Duration) ServerOpt {
	return func(config *ServerConfig) {
		config.ReadTimeout = timeout
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

// TrustedProxies is a list of CIDR address ranges that can be trusted when handling RemoteIP header.
func TrustedProxies(trustedProxies []string) ServerOpt {
	return func(config *ServerConfig) {
		config.TrustedProxies = trustedProxies
	}
}

// RemoteIPHeader is a name of the header that overwrites the value of client's remote address.
func RemoteIPHeader(header string) ServerOpt {
	return func(config *ServerConfig) {
		config.RemoteIPHeader = header
	}
}
