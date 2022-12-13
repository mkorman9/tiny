package tinyredis

import (
	"crypto/tls"
	"time"
)

// Config holds a configuration for Dial call.
type Config struct {
	// Username is an optional property used in authorization.
	Username string

	// Password is an optional property used in authorization.
	Password string

	// DB is a database number to use (default: 0).
	DB int

	// TLSConfig setting it to non-nil value enables TLS mode.
	TLSConfig *tls.Config

	// ConnectionTimeout is a maximum time client should spend trying to connect (default: 5s).
	ConnectionTimeout time.Duration

	// NoPing indicates whether Dial should skip the initial call to Ping method (default: false).
	NoPing bool
}

// Opt is an option to be specified to Dial.
type Opt = func(*Config)

// Username is an optional property used in authorization.
func Username(username string) Opt {
	return func(config *Config) {
		config.Username = username
	}
}

// Password is an optional property used in authorization.
func Password(password string) Opt {
	return func(config *Config) {
		config.Password = password
	}
}

// DB is a database number to use.
func DB(db int) Opt {
	return func(config *Config) {
		config.DB = db
	}
}

// TLSConfig setting it to non-nil value enables TLS mode.
func TLSConfig(tlsConfig *tls.Config) Opt {
	return func(config *Config) {
		config.TLSConfig = tlsConfig
	}
}

// ConnectionTimeout is a maximum time client should spend trying to connect.
func ConnectionTimeout(connectionTimeout time.Duration) Opt {
	return func(config *Config) {
		config.ConnectionTimeout = connectionTimeout
	}
}

// NoPing indicates that Dial should skip the initial call to Ping method.
func NoPing() Opt {
	return func(config *Config) {
		config.NoPing = true
	}
}
