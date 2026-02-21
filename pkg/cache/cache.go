package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache provides a simple Redis caching interface
type Cache struct {
	client *redis.Client
}

// New creates a new Cache instance. If client is nil, all operations are no-ops.
func New(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// Enabled returns true if Redis is available
func (c *Cache) Enabled() bool {
	return c.client != nil
}

// Get retrieves a cached value and unmarshals it into dest.
// Returns true if cache hit, false if miss or error.
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) bool {
	if c.client == nil {
		return false
	}

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return false
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false
	}

	return true
}

// Set stores a value in cache with the given TTL.
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	if c.client == nil {
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	c.client.Set(ctx, key, data, ttl)
}

// Delete removes specific keys from cache.
func (c *Cache) Delete(ctx context.Context, keys ...string) {
	if c.client == nil {
		return
	}

	c.client.Del(ctx, keys...)
}

// DeleteByPattern removes all keys matching a glob pattern (e.g. "user:abc:*").
func (c *Cache) DeleteByPattern(ctx context.Context, pattern string) {
	if c.client == nil {
		return
	}

	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if len(keys) > 0 {
		c.client.Del(ctx, keys...)
	}
}
