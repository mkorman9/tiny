package tinypostgres

import (
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Client is a wrapper for *gorm.DB providing a handy Close() function.
type Client struct {
	*gorm.DB
}

// DialPostgres creates a connection to Postgres, and returns Client instance.
func DialPostgres(opts ...Opt) (*Client, error) {
	config := Config{
		Verbose:         false,
		PoolMaxOpen:     10,
		PoolMaxIdle:     5,
		PoolMaxLifetime: time.Hour,
		PoolMaxIdleTime: 30 * time.Minute,
	}

	for _, opt := range opts {
		opt(&config)
	}

	if config.DSN == "" {
		return nil, errors.New("DSN cannot be empty")
	}

	gormConfig := &gorm.Config{
		Logger:      &gormLogger{verbose: config.Verbose},
		NowFunc:     func() time.Time { return time.Now().UTC() },
		QueryFields: true,
	}

	for _, opt := range config.gormOpts {
		opt(gormConfig)
	}

	log.Debug().Msg("Establishing Postgres connection")

	db, err := gorm.Open(postgres.Open(config.DSN), gormConfig)
	if err == nil {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}

		sqlDB.SetMaxOpenConns(config.PoolMaxOpen)
		sqlDB.SetMaxIdleConns(config.PoolMaxIdle)
		sqlDB.SetConnMaxLifetime(config.PoolMaxLifetime)
		sqlDB.SetConnMaxIdleTime(config.PoolMaxIdleTime)

		log.Info().Msg("Successfully connected to Postgres")
	}

	return &Client{DB: db}, err
}

// Close closes a connection to Postgres.
func (c *Client) Close() {
	log.Debug().Msg("Closing Postgres connection")

	sqlDB, err := c.DB.DB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire *sql.DB reference when closing Postgres connection")
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Error().Err(err).Msg("Error when closing Postgres connection")
	} else {
		log.Info().Msg("Postgres connection closed successfully")
	}
}
