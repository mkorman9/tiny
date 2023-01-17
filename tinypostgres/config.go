package tinypostgres

import (
	"gorm.io/gorm"
	"time"
)

// Config holds a configuration for Dial.
type Config struct {
	// Verbose specifies whether to log all executed queries.
	Verbose bool

	// PoolMaxOpen is the maximum number of open connections to the database (default: 10).
	PoolMaxOpen int

	// PoolMaxIdle is the maximum number of connections in the idle connection pool (default: 5).
	PoolMaxIdle int

	// PoolMaxLifetime is the maximum amount of time a connection may be reused (default: 1h).
	PoolMaxLifetime time.Duration

	// PoolMaxIdleTime is the maximum amount of time a connection may be idle (default: 30m).
	PoolMaxIdleTime time.Duration

	// GormOpt allows to specify custom function that will operate directly on *gorm.Config.
	GormOpt func(*gorm.Config)
}

func mergeConfig(provided *Config) *Config {
	config := &Config{
		PoolMaxOpen:     10,
		PoolMaxIdle:     5,
		PoolMaxLifetime: time.Hour,
		PoolMaxIdleTime: 30 * time.Minute,
	}

	if provided == nil {
		return config
	}

	if provided.Verbose {
		config.Verbose = true
	}
	if provided.PoolMaxOpen > 0 {
		config.PoolMaxOpen = provided.PoolMaxOpen
	}
	if provided.PoolMaxIdle > 0 {
		config.PoolMaxIdle = provided.PoolMaxIdle
	}
	if provided.PoolMaxIdleTime > 0 {
		config.PoolMaxIdleTime = provided.PoolMaxIdleTime
	}
	if provided.PoolMaxIdleTime > 0 {
		config.PoolMaxIdleTime = provided.PoolMaxIdleTime
	}
	if provided.GormOpt != nil {
		config.GormOpt = provided.GormOpt
	}

	return config
}
