package config_test

import (
	"context"
	"testing"
	"time"

	"hifzhun-api/pkg/config"

	gormlogger "gorm.io/gorm/logger"
)

func TestGormSlowQueryLogger(t *testing.T) {
	logger := config.NewGormSlowQueryLogger(100 * time.Millisecond)

	// Test LogMode
	newLogger := logger.LogMode(gormlogger.Info)
	if newLogger == nil {
		t.Fatal("expected non-nil logger after LogMode")
	}

	// Test Info, Warn, Error
	ctx := context.Background()
	logger.Info(ctx, "test info: %s", "hello")
	logger.Warn(ctx, "test warn: %s", "hello")
	logger.Error(ctx, "test error: %s", "hello")

	// Test Trace (Fast query - should not trigger slow query log)
	logger.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	// Test Trace (Slow query - triggers slow query log)
	logger.Trace(ctx, time.Now().Add(-150*time.Millisecond), func() (string, int64) {
		return "SELECT * FROM items WHERE status = 'interval'", 50
	}, nil)

	// Test Trace (Sensitive SQL masking)
	logger.Trace(ctx, time.Now().Add(-150*time.Millisecond), func() (string, int64) {
		return "SELECT * FROM users WHERE password = 'secretpassword'", 1
	}, nil)
}
