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

	gormOpts []func(*gorm.Config)
}

// Opt is an option to be specified to Dial.
type Opt = func(*Config)

// Verbose tells client to log all executed queries.
func Verbose() Opt {
	return func(config *Config) {
		config.Verbose = true
	}
}

// PoolMaxOpen is the maximum number of open connections to the database.
func PoolMaxOpen(poolMaxOpen int) Opt {
	return func(config *Config) {
		config.PoolMaxOpen = poolMaxOpen
	}
}

// PoolMaxIdle is the maximum number of connections in the idle connection pool.
func PoolMaxIdle(poolMaxIdle int) Opt {
	return func(config *Config) {
		config.PoolMaxIdle = poolMaxIdle
	}
}

// PoolMaxLifetime is the maximum amount of time a connection may be reused.
func PoolMaxLifetime(poolMaxLifetime time.Duration) Opt {
	return func(config *Config) {
		config.PoolMaxLifetime = poolMaxLifetime
	}
}

// PoolMaxIdleTime is the maximum amount of time a connection may be idle.
func PoolMaxIdleTime(poolMaxIdleTime time.Duration) Opt {
	return func(config *Config) {
		config.PoolMaxIdleTime = poolMaxIdleTime
	}
}

// GormOpt adds an option to modify the default gorm.Config.
func GormOpt(gormOpt func(*gorm.Config)) Opt {
	return func(config *Config) {
		config.gormOpts = append(config.gormOpts, gormOpt)
	}
}
