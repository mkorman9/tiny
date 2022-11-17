package tinyredis

import (
	"crypto/tls"
	"time"
)

// Config holds a configuration for Client.
type Config struct {
	// Address is a remote host and port to connect to.
	Address string

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
}

// Opt is an option to be specified to DialRedis.
type Opt = func(*Config)

// Address is a remote host and port to connect to.
func Address(address string) Opt {
	return func(config *Config) {
		config.Address = address
	}
}

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
