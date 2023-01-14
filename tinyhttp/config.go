package tinyhttp

import (
	"crypto/tls"
	"github.com/gofiber/fiber/v2"
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

	// TLSConfig is an optional TLS configuration to pass when using TLS mode.
	TLSConfig *tls.Config

	// ReadTimeout is a timeout used when creating underlying http server (default: 5s).
	ReadTimeout time.Duration

	// WriteTimeout is a timeout used when creating underlying http server (default: 10s).
	WriteTimeout time.Duration

	// IdleTimeout is a timeout used when creating underlying http server (default 2m).
	IdleTimeout time.Duration

	// TrustedProxies is a list of CIDR address ranges that can be trusted when handling RemoteIP header.
	// (default: "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8", "fc00::/7", "::1/128")
	TrustedProxies []string

	// RemoteIPHeaders is a list of headers that overwrite the value of client's remote address.
	// (default: "X-Forwarded-For")
	RemoteIPHeader string

	// ViewEngine is a template rendering engine for fiber (default: nil).
	ViewEngine fiber.Views

	// ViewLayout is a global layout for ViewEngine (default: "").
	ViewLayout string

	// Concurrency specifies a maximum number of concurrent connections (default: 256 * 1024).
	Concurrency int

	// BodyLimit specifies a maximum allowed size for a request body (default: 4 * 1024 * 1024).
	BodyLimit int

	// ReadBufferSize specifies a per-connection buffer size (default: 4096).
	ReadBufferSize int

	// WriteBufferSize specifies a per-connection buffer size for responses (default: 4096).
	WriteBufferSize int

	fiberOption func(*fiber.Config)
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
func TLS(cert, key string, tlsConfig ...*tls.Config) ServerOpt {
	return func(config *ServerConfig) {
		config.TLSCert = cert
		config.TLSKey = key

		if tlsConfig != nil {
			config.TLSConfig = tlsConfig[0]
		}
	}
}

// ReadTimeout is a timeout used when creating underlying http server.
func ReadTimeout(timeout time.Duration) ServerOpt {
	return func(config *ServerConfig) {
		config.ReadTimeout = timeout
	}
}

// WriteTimeout is a timeout used when creating underlying http server.
func WriteTimeout(timeout time.Duration) ServerOpt {
	return func(config *ServerConfig) {
		config.WriteTimeout = timeout
	}
}

// IdleTimeout is a timeout used when creating underlying http server.
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

// FiberOption specifies user-defined function that directly modifies fiber config.
func FiberOption(opt func(*fiber.Config)) ServerOpt {
	return func(config *ServerConfig) {
		config.fiberOption = opt
	}
}

// ViewEngine is a template rendering engine for fiber.
func ViewEngine(engine fiber.Views) ServerOpt {
	return func(config *ServerConfig) {
		config.ViewEngine = engine
	}
}

// ViewLayout is a global layout for ViewEngine.
func ViewLayout(layout string) ServerOpt {
	return func(config *ServerConfig) {
		config.ViewLayout = layout
	}
}

// Concurrency specifies a maximum number of concurrent connections.
func Concurrency(concurrency int) ServerOpt {
	return func(config *ServerConfig) {
		config.Concurrency = concurrency
	}
}

// BodyLimit specifies a maximum allowed size for a request body.
func BodyLimit(bodyLimit int) ServerOpt {
	return func(config *ServerConfig) {
		config.BodyLimit = bodyLimit
	}
}

// ReadBufferSize specifies a per-connection buffer size.
func ReadBufferSize(readBufferSize int) ServerOpt {
	return func(config *ServerConfig) {
		config.ReadBufferSize = readBufferSize
	}
}

// WriteBufferSize specifies a per-connection buffer size for responses.
func WriteBufferSize(writeBufferSize int) ServerOpt {
	return func(config *ServerConfig) {
		config.WriteBufferSize = writeBufferSize
	}
}
