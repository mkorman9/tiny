package tinyredis

import (
	"github.com/go-redis/redis/v8"
	"time"
)

// Config holds a configuration for Dial call.
type Config struct {
	// ConnectionTimeout is a maximum time client should spend trying to connect (default: 5s).
	ConnectionTimeout time.Duration

	// NoPing indicates whether Dial should skip the initial call to Ping method (default: false).
	NoPing bool

	redisOpts func(*redis.Options)
}

// Opt is an option to be specified to Dial.
type Opt = func(*Config)

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

// Options sets a function that allows to customize redis options.
func Options(redisOpt func(options *redis.Options)) Opt {
	return func(config *Config) {
		config.redisOpts = redisOpt
	}
}
