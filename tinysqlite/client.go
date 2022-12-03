package tinysqlite

import (
	"errors"
	"github.com/glebarez/sqlite"
	"github.com/mkorman9/tiny/gormcommon"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"time"
)

// Client is a wrapper for *gorm.DB providing a handy Close() function.
type Client struct {
	*gorm.DB
}

// Open tries to open an instance of sqlite3 database and then return Client instance to interact with it.
func Open(dsn string, opts ...Opt) (*Client, error) {
	config := Config{
		Verbose: false,
	}

	for _, opt := range opts {
		opt(&config)
	}

	if dsn == "" {
		return nil, errors.New("DSN cannot be empty")
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

	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	return &Client{DB: db}, err
}

// Close closes a connection to Postgres.
func (c *Client) Close() {
	sqlDB, err := c.DB.DB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire *sql.DB reference when closing sqlite database")
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Error().Err(err).Msg("Error when closing sqlite database")
	}
}
