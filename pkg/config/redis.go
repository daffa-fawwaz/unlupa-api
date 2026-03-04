package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	// Allow disabling cache entirely via env
	if os.Getenv("CACHE_DISABLED") == "true" {
		log.Printf("⚠️  Cache disabled via env (CACHE_DISABLED=true)")
		RedisClient = nil
		return
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Printf("⚠️  REDIS_URL not set, caching disabled")
		RedisClient = nil
		return
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  Redis connection failed: %v (caching disabled)", err)
		RedisClient = nil
		return
	}

	log.Println("✅ Redis connected successfully!")
}
