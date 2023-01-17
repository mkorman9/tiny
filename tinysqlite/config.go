package tinysqlite

import (
	"gorm.io/gorm"
)

// Config holds a configuration for Open.
type Config struct {
	// Verbose specifies whether to log all executed queries.
	Verbose bool

	// GormOpt allows to specify custom function that will operate directly on *gorm.Config.
	GormOpt func(*gorm.Config)
}

func mergeConfig(provided *Config) *Config {
	config := &Config{}

	if provided == nil {
		return config
	}

	if provided.Verbose {
		config.Verbose = true
	}
	if provided.GormOpt != nil {
		config.GormOpt = provided.GormOpt
	}

	return config
}
