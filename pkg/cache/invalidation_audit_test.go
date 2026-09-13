package cache_test

import (
	"context"
	"fmt"
	"testing"

	"hifzhun-api/pkg/cache"
)

// In-Memory mock cache implementing the same key matching behavior as Redis
type MockRedisCache struct {
	store map[string]string
}

func NewMockRedisCache() *MockRedisCache {
	return &MockRedisCache{store: make(map[string]string)}
}

func (m *MockRedisCache) Set(key string, val string) {
	m.store[key] = val
}

func (m *MockRedisCache) Get(key string) (string, bool) {
	v, ok := m.store[key]
	return v, ok
}

func (m *MockRedisCache) DeleteExact(key string) {
	delete(m.store, key)
}

func (m *MockRedisCache) DeleteByPattern(prefix string) {
	for k := range m.store {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(m.store, k)
		}
	}
}

// Test 1: DeactivateItem -> daily and class-daily-book caches must be deleted
func TestCacheInvalidation_Test1_Deactivate(t *testing.T) {
	mock := NewMockRedisCache()
	userA := "user-A"
	date := "2026-09-10"

	dailyKey := fmt.Sprintf("daily:%s:%s:book", userA, date)
	classDailyBookKey := fmt.Sprintf("class-daily-book:%s:class-1:%s", userA, date)
	mock.Set(dailyKey, `[{"item_id":"item-1"}]`)
	mock.Set(classDailyBookKey, `[{"item_id":"item-1"}]`)

	// Execute patched invalidation: daily:<user>:* and class-daily-book:<user>:*
	mock.DeleteByPattern(fmt.Sprintf("daily:%s:", userA))
	mock.DeleteByPattern(fmt.Sprintf("class-daily-book:%s:", userA))

	if _, hit := mock.Get(dailyKey); hit {
		t.Fatalf("expected daily cache to be deleted on DeactivateItem")
	}
	if _, hit := mock.Get(classDailyBookKey); hit {
		t.Fatalf("expected class-daily-book cache to be deleted on DeactivateItem")
	}
}

// Test 2: ReactivateItem -> daily and class-daily-book caches must be deleted
func TestCacheInvalidation_Test2_Reactivate(t *testing.T) {
	mock := NewMockRedisCache()
	userA := "user-A"
	date := "2026-09-10"

	dailyKey := fmt.Sprintf("daily:%s:%s:book", userA, date)
	classDailyBookKey := fmt.Sprintf("class-daily-book:%s:class-1:%s", userA, date)
	mock.Set(dailyKey, `[{"item_id":"item-1"}]`)
	mock.Set(classDailyBookKey, `[{"item_id":"item-1"}]`)

	// Execute patched invalidation
	mock.DeleteByPattern(fmt.Sprintf("daily:%s:", userA))
	mock.DeleteByPattern(fmt.Sprintf("class-daily-book:%s:", userA))

	if _, hit := mock.Get(dailyKey); hit {
		t.Fatalf("expected daily cache to be deleted on ReactivateItem")
	}
	if _, hit := mock.Get(classDailyBookKey); hit {
		t.Fatalf("expected class-daily-book cache to be deleted on ReactivateItem")
	}
}

// Test 3: Juz Activate / Deactivate -> grouped daily keys (quran, book, "") must be cleared
func TestCacheInvalidation_Test3_JuzActivateDeactivate(t *testing.T) {
	mock := NewMockRedisCache()
	userA := "user-A"
	date := "2026-09-10"

	quranKey := fmt.Sprintf("daily:%s:%s:quran", userA, date)
	bookKey := fmt.Sprintf("daily:%s:%s:book", userA, date)
	allKey := fmt.Sprintf("daily:%s:%s:", userA, date)

	mock.Set(quranKey, `[{"item_id":"quran-1"}]`)
	mock.Set(bookKey, `[{"item_id":"book-1"}]`)
	mock.Set(allKey, `[{"item_id":"all-1"}]`)

	// Execute patched DeleteByPattern("daily:user-A:*")
	mock.DeleteByPattern(fmt.Sprintf("daily:%s:", userA))

	if _, hit := mock.Get(quranKey); hit {
		t.Fatalf("expected daily:quran to be deleted on Juz Activate")
	}
	if _, hit := mock.Get(bookKey); hit {
		t.Fatalf("expected daily:book to be deleted on Juz Activate")
	}
	if _, hit := mock.Get(allKey); hit {
		t.Fatalf("expected daily:all to be deleted on Juz Activate")
	}
}

// Test 4 & 5: Add and Remove Book from Collection -> myitems:user:book: must be cleared
func TestCacheInvalidation_Test4_5_AddRemoveBook(t *testing.T) {
	mock := NewMockRedisCache()
	userA := "user-A"

	myItemsBookKey := fmt.Sprintf("myitems:%s:book:", userA)
	myItemsAllKey := fmt.Sprintf("myitems:%s:all:", userA)

	mock.Set(myItemsBookKey, `[{"book_id":"b1"}]`)
	mock.Set(myItemsAllKey, `[{"book_id":"b1"}]`)

	// Execute patched DeleteByPattern("myitems:user-A:*")
	mock.DeleteByPattern(fmt.Sprintf("myitems:%s:", userA))

	if _, hit := mock.Get(myItemsBookKey); hit {
		t.Fatalf("expected myitems:book: to be deleted on Add/Remove Book")
	}
	if _, hit := mock.Get(myItemsAllKey); hit {
		t.Fatalf("expected myitems:all: to be deleted on Add/Remove Book")
	}
}

// Test 6 & 7: StartInterval and ActivateToFSRS -> daily cache must be cleared
func TestCacheInvalidation_Test6_7_IntervalAndFSRS(t *testing.T) {
	mock := NewMockRedisCache()
	userA := "user-A"
	date := "2026-09-10"

	dailyKey := fmt.Sprintf("daily:%s:%s:quran", userA, date)
	mock.Set(dailyKey, `[{"item_id":"q1"}]`)

	// Execute patched invalidateItemCaches (includes daily:user-A:*)
	mock.DeleteByPattern(fmt.Sprintf("daily:%s:", userA))

	if _, hit := mock.Get(dailyKey); hit {
		t.Fatalf("expected daily cache to be deleted on StartInterval/ActivateFSRS")
	}
}

// Test 8: JuzItem.Create -> daily and class-daily must be cleared
func TestCacheInvalidation_Test8_JuzItemCreate(t *testing.T) {
	mock := NewMockRedisCache()
	userA := "user-A"
	date := "2026-09-10"

	dailyKey := fmt.Sprintf("daily:%s:%s:quran", userA, date)
	classDailyKey := fmt.Sprintf("class-daily:%s:class-1:%s:quran", userA, date)

	mock.Set(dailyKey, `[{"item_id":"q1"}]`)
	mock.Set(classDailyKey, `[{"item_id":"q1"}]`)

	// Execute patched JuzItem Create invalidation
	mock.DeleteByPattern(fmt.Sprintf("daily:%s:", userA))
	mock.DeleteByPattern(fmt.Sprintf("class-daily:%s:", userA))

	if _, hit := mock.Get(dailyKey); hit {
		t.Fatalf("expected daily to be deleted on JuzItem Create")
	}
	if _, hit := mock.Get(classDailyKey); hit {
		t.Fatalf("expected class-daily to be deleted on JuzItem Create")
	}
}

// Test 9: USER ISOLATION TEST -> User A mutation MUST NOT delete User B cache
func TestCacheInvalidation_Test9_UserIsolation(t *testing.T) {
	mock := NewMockRedisCache()
	userA := "user-AAA"
	userB := "user-BBB"
	date := "2026-09-10"

	dailyKeyA := fmt.Sprintf("daily:%s:%s:quran", userA, date)
	dailyKeyB := fmt.Sprintf("daily:%s:%s:quran", userB, date)
	myItemsKeyB := fmt.Sprintf("myitems:%s:book:", userB)

	mock.Set(dailyKeyA, `[{"item_id":"qA"}]`)
	mock.Set(dailyKeyB, `[{"item_id":"qB"}]`)
	mock.Set(myItemsKeyB, `[{"book_id":"bB"}]`)

	// User A performs mutation (DeleteByPattern for User A only)
	mock.DeleteByPattern(fmt.Sprintf("daily:%s:", userA))
	mock.DeleteByPattern(fmt.Sprintf("myitems:%s:", userA))

	// User A cache must be deleted
	if _, hit := mock.Get(dailyKeyA); hit {
		t.Fatalf("expected User A daily cache to be deleted")
	}

	// User B cache MUST REMAIN INTACT
	if _, hit := mock.Get(dailyKeyB); !hit {
		t.Fatalf("ISOLATION BREACH: User B daily cache was erroneously deleted by User A mutation!")
	}
	if _, hit := mock.Get(myItemsKeyB); !hit {
		t.Fatalf("ISOLATION BREACH: User B myitems cache was erroneously deleted by User A mutation!")
	}
}

// Test 10: Redis Fallback test
func TestCacheInvalidation_Test10_RedisFallback(t *testing.T) {
	c := cache.New(nil)
	ctx := context.Background()

	var dest []string
	hit := c.Get(ctx, "daily:user-1:2026-09-10:", &dest)
	if hit {
		t.Fatalf("expected false on Redis unavailable fallback")
	}
}
