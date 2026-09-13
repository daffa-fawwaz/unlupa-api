package config

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// GormSlowQueryLogger implements gormlogger.Interface to log slow queries in a structured format
type GormSlowQueryLogger struct {
	SlowThreshold time.Duration
	LogLevel      gormlogger.LogLevel
}

func NewGormSlowQueryLogger(slowThreshold time.Duration) *GormSlowQueryLogger {
	return &GormSlowQueryLogger{
		SlowThreshold: slowThreshold,
		LogLevel:      gormlogger.Warn,
	}
}

func (l *GormSlowQueryLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GormSlowQueryLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		log.Printf("[DB_INFO] "+msg, data...)
	}
}

func (l *GormSlowQueryLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		log.Printf("[DB_WARN] "+msg, data...)
	}
}

func (l *GormSlowQueryLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		log.Printf("[DB_ERROR] "+msg, data...)
	}
}

func (l *GormSlowQueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	durationMs := elapsed.Milliseconds()

	// Mask sensitive information in SQL if any
	maskSensitiveSQL := func(rawSQL string) string {
		lower := strings.ToLower(rawSQL)
		if strings.Contains(lower, "password") || strings.Contains(lower, "secret") || strings.Contains(lower, "jwt") {
			return "[REDACTED_SENSITIVE_SQL]"
		}
		// Truncate SQL if excessively long (> 1000 chars) to prevent log flooding
		if len(rawSQL) > 1000 {
			return rawSQL[:1000] + "... [truncated]"
		}
		return rawSQL
	}

	source := utils.FileWithLineNum()

	if err != nil && l.LogLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound) {
		sql, rows := fc()
		log.Printf("[DB_ERROR] duration_ms=%d rows=%d source=%s err=%v sql=\"%s\"",
			durationMs, rows, source, err, maskSensitiveSQL(sql))
		return
	}

	if l.SlowThreshold != 0 && elapsed >= l.SlowThreshold && l.LogLevel >= gormlogger.Warn {
		sql, rows := fc()
		log.Printf("[SLOW_QUERY] duration_ms=%d rows=%d source=%s sql=\"%s\"",
			durationMs, rows, source, maskSensitiveSQL(sql))
		return
	}
}
