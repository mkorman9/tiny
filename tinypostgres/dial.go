package tinypostgres

import (
	"errors"
	"github.com/mkorman9/tiny/gormcommon"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Dial creates a connection to Postgres, and returns *gorm.DB instance.
func Dial(url string, config ...*Config) (*gorm.DB, error) {
	var providedConfig *Config
	if config != nil {
		providedConfig = config[0]
	}
	c := mergeConfig(providedConfig)

	if url == "" {
		return nil, errors.New("URL cannot be empty")
	}

	gormConfig := &gorm.Config{
		Logger: &gormcommon.GormLogger{Verbose: c.Verbose},
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		QueryFields: true,
	}

	if c.GormOpt != nil {
		c.GormOpt(gormConfig)
	}

	db, err := gorm.Open(postgres.Open(url), gormConfig)
	if err == nil {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}

		sqlDB.SetMaxOpenConns(c.PoolMaxOpen)
		sqlDB.SetMaxIdleConns(c.PoolMaxIdle)
		sqlDB.SetConnMaxLifetime(c.PoolMaxLifetime)
		sqlDB.SetConnMaxIdleTime(c.PoolMaxIdleTime)
	}

	return db, err
}
