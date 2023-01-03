package tinysqlite

import (
	"errors"
	"github.com/glebarez/sqlite"
	"github.com/mkorman9/tiny/gormcommon"
	"gorm.io/gorm"
	"time"
)

// Open tries to open an instance of sqlite3 database and then return *gorm.DB instance to interact with it.
func Open(dsn string, opts ...Opt) (*gorm.DB, error) {
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

	return db, err
}
