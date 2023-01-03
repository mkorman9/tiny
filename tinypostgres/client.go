package tinypostgres

import (
	"errors"
	"github.com/mkorman9/tiny/gormcommon"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Dial creates a connection to Postgres, and returns *gorm.DB instance.
func Dial(url string, opts ...Opt) (*gorm.DB, error) {
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

	return db, err
}
