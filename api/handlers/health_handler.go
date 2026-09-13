package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"hifzhun-api/pkg/config"
)

func Health(c *fiber.Ctx) error {
	dbOK := false
	redisOK := false

	if config.DB != nil {
		if sqlDB, err := config.DB.DB(); err == nil {
			if err := sqlDB.Ping(); err == nil {
				dbOK = true
			}
		}
	}

	if config.RedisClient != nil {
		if err := config.RedisClient.Ping(context.Background()).Err(); err == nil {
			redisOK = true
		}
	}

	res := fiber.Map{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
		"db":     dbOK,
		"redis":  redisOK,
	}

	if poolStats, ok := config.GetDBPoolStats(); ok {
		res["db_pool"] = poolStats
	}

	return c.JSON(res)
}
