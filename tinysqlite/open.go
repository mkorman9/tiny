package tinysqlite

import (
	"errors"
	"github.com/glebarez/sqlite"
	"github.com/mkorman9/tiny/gormcommon"
	"gorm.io/gorm"
	"time"
)

// Open tries to open an instance of sqlite3 database and then return *gorm.DB instance to interact with it.
func Open(dsn string, config ...*Config) (*gorm.DB, error) {
	var providedConfig *Config
	if config != nil {
		providedConfig = config[0]
	}
	c := mergeConfig(providedConfig)

	if dsn == "" {
		return nil, errors.New("DSN cannot be empty")
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

	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	return db, err
}
