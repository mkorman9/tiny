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

	// RedisOpt allows to specify a function that operates directly on *redis.Options.
	RedisOpt func(*redis.Options)
}

func mergeConfig(provided *Config) *Config {
	config := &Config{
		ConnectionTimeout: 5 * time.Second,
	}

	if provided == nil {
		return config
	}

	if provided.ConnectionTimeout > 0 {
		config.ConnectionTimeout = provided.ConnectionTimeout
	}
	if provided.NoPing {
		config.NoPing = true
	}
	if provided.RedisOpt != nil {
		config.RedisOpt = provided.RedisOpt
	}

	return config
}
