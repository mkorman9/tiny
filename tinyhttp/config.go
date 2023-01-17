package tinyhttp

import (
	"crypto/tls"
	"github.com/gofiber/fiber/v2"
	"time"
)

// ServerConfig holds a configuration for NewServer.
type ServerConfig struct {
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

	// FiberOpt allows to specify custom function that will operate directly on *fiber.Config.
	FiberOpt func(*fiber.Config)
}

func mergeServerConfig(provided *ServerConfig) *ServerConfig {
	config := &ServerConfig{
		Network:         "tcp",
		SecurityHeaders: true,
		ShutdownTimeout: 5 * time.Second,
		TLSConfig:       &tls.Config{},
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     2 * time.Minute,
		TrustedProxies: []string{
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			"127.0.0.0/8",
			"fc00::/7",
			"::1/128",
		},
		RemoteIPHeader:  "X-Forwarded-For",
		Concurrency:     256 * 1024,
		BodyLimit:       4 * 1024 * 1024,
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}

	if provided == nil {
		return config
	}

	if provided.Network != "" {
		config.Network = provided.Network
	}
	if provided.SecurityHeaders {
		config.SecurityHeaders = true
	}
	if provided.ShutdownTimeout > 0 {
		config.ShutdownTimeout = provided.ShutdownTimeout
	}
	if provided.TLSCert != "" {
		config.TLSCert = provided.TLSCert
	}
	if provided.TLSKey != "" {
		config.TLSKey = provided.TLSKey
	}
	if provided.TLSConfig != nil {
		config.TLSConfig = provided.TLSConfig
	}
	if provided.ReadTimeout > 0 {
		config.ReadTimeout = provided.ReadTimeout
	}
	if provided.WriteTimeout > 0 {
		config.WriteTimeout = provided.WriteTimeout
	}
	if provided.IdleTimeout > 0 {
		config.IdleTimeout = provided.IdleTimeout
	}
	if provided.TrustedProxies != nil {
		config.TrustedProxies = provided.TrustedProxies
	}
	if provided.RemoteIPHeader != "" {
		config.RemoteIPHeader = provided.RemoteIPHeader
	}
	if provided.ViewEngine != nil {
		config.ViewEngine = provided.ViewEngine
	}
	if provided.ViewLayout != "" {
		config.ViewLayout = provided.ViewLayout
	}
	if provided.Concurrency > 0 {
		config.Concurrency = provided.Concurrency
	}
	if provided.BodyLimit > 0 {
		config.BodyLimit = provided.BodyLimit
	}
	if provided.ReadBufferSize > 0 {
		config.ReadBufferSize = provided.ReadBufferSize
	}
	if provided.WriteBufferSize > 0 {
		config.WriteBufferSize = provided.WriteBufferSize
	}
	if provided.FiberOpt != nil {
		config.FiberOpt = provided.FiberOpt
	}

	return config
}
