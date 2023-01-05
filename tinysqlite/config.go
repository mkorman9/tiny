package tinysqlite

import (
	"gorm.io/gorm"
)

// Config holds a configuration for Open.
type Config struct {
	// Verbose specifies whether to log all executed queries.
	Verbose bool

	gormOpts []func(*gorm.Config)
}

// Opt is an option to be specified to DialPostgres.
type Opt = func(*Config)

// Verbose tells client to log all executed queries.
func Verbose() Opt {
	return func(config *Config) {
		config.Verbose = true
	}
}

// GormOpt adds an option to modify the default gorm.Config.
func GormOpt(gormOpt func(*gorm.Config)) Opt {
	return func(config *Config) {
		config.gormOpts = append(config.gormOpts, gormOpt)
	}
}
