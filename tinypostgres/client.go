package tinypostgres

import (
	"errors"
	"github.com/mkorman9/tiny/gormcommon"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Client is a wrapper for *gorm.DB providing a handy Close() function.
type Client struct {
	*gorm.DB
}

// Dial creates a connection to Postgres, and returns Client instance.
func Dial(url string, opts ...Opt) (*Client, error) {
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

	if url == "" {
		return nil, errors.New("URL cannot be empty")
	}

	gormConfig := &gorm.Config{
		Logger: &gormcommon.GormLogger{Verbose: config.Verbose},
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		QueryFields: true,
	}

	for _, opt := range config.gormOpts {
		opt(gormConfig)
	}

	db, err := gorm.Open(postgres.Open(url), gormConfig)
	if err == nil {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}

		sqlDB.SetMaxOpenConns(config.PoolMaxOpen)
		sqlDB.SetMaxIdleConns(config.PoolMaxIdle)
		sqlDB.SetConnMaxLifetime(config.PoolMaxLifetime)
		sqlDB.SetConnMaxIdleTime(config.PoolMaxIdleTime)
	}

	return &Client{DB: db}, err
}

// Close closes a connection to Postgres.
func (c *Client) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
