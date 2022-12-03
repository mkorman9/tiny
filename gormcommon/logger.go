package gormcommon

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"
)

type GormLogger struct {
	Verbose bool
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	log.Info().Msgf(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	log.Warn().Msgf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	log.Error().Msgf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	query, rows := fc()

	if err != nil {
		log.Warn().Err(err).Msgf("DB error for: '%s'", query)
	} else if l.Verbose {
		elapsed := time.Now().UTC().Sub(begin)
		log.Debug().Msgf("DB query (%v) [%d rows]: '%s'", elapsed.String(), rows, query)
	}
}
