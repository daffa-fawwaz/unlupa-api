package cache_test

import (
	"context"
	"testing"
	"time"

	"hifzhun-api/pkg/cache"
)

type DummyResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func TestCache_NilClientFallback(t *testing.T) {
	// When Redis client is nil (CACHE_DISABLED=true or connection failed)
	c := cache.New(nil)

	if c.Enabled() {
		t.Errorf("expected Enabled() to be false when client is nil")
	}

	ctx := context.Background()

	// 1. Get should safely return false (CACHE MISS / FALLBACK) without error/panic
	var dest DummyResponse
	if hit := c.Get(ctx, "test:key", &dest); hit {
		t.Errorf("expected Get to return false when client is nil")
	}

	// 2. Set should be a safe no-op without error/panic
	c.Set(ctx, "test:key", DummyResponse{ID: "1", Title: "Test"}, 10*time.Minute)

	// 3. Delete should be a safe no-op without error/panic
	c.Delete(ctx, "test:key")

	// 4. DeleteByPattern should be a safe no-op without error/panic
	c.DeleteByPattern(ctx, "test:*")
}
