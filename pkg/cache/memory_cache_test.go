package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_SetGet(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Set value
	err := cache.Set(ctx, "key1", "value1", TTLShort)
	assert.NoError(t, err)

	// Get value
	var result string
	err = cache.Get(ctx, "key1", &result)
	assert.NoError(t, err)
	assert.Equal(t, "value1", result)
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	var result string
	err := cache.Get(ctx, "nonexistent", &result)
	assert.Error(t, err)
}

func TestMemoryCache_SetStruct(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	type TestData struct {
		Name string
		Age  int
	}

	// Set struct
	original := TestData{Name: "John", Age: 30}
	err := cache.Set(ctx, "user:1", original, TTLShort)
	assert.NoError(t, err)

	// Get struct
	var result TestData
	err = cache.Get(ctx, "user:1", &result)
	assert.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Set and delete
	cache.Set(ctx, "key1", "value1", TTLShort)
	err := cache.Delete(ctx, "key1")
	assert.NoError(t, err)

	// Verify deleted
	var result string
	err = cache.Get(ctx, "key1", &result)
	assert.Error(t, err)
}

func TestMemoryCache_DeletePattern(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Set multiple keys
	cache.Set(ctx, "user:1", "value1", TTLShort)
	cache.Set(ctx, "user:2", "value2", TTLShort)
	cache.Set(ctx, "post:1", "value3", TTLShort)

	// Delete pattern
	err := cache.DeletePattern(ctx, "user:*")
	assert.NoError(t, err)

	// Verify user keys deleted
	var result string
	err = cache.Get(ctx, "user:1", &result)
	assert.Error(t, err)

	err = cache.Get(ctx, "user:2", &result)
	assert.Error(t, err)

	// Verify post key still exists
	err = cache.Get(ctx, "post:1", &result)
	assert.NoError(t, err)
}

func TestMemoryCache_Exists(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Should not exist
	exists, err := cache.Exists(ctx, "key1")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Set value
	cache.Set(ctx, "key1", "value1", TTLShort)

	// Should exist
	exists, err = cache.Exists(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Set with short TTL
	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	assert.NoError(t, err)

	// Should exist immediately
	exists, _ := cache.Exists(ctx, "key1")
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not exist anymore
	exists, _ = cache.Exists(ctx, "key1")
	assert.False(t, exists)
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Set multiple keys
	cache.Set(ctx, "key1", "value1", TTLShort)
	cache.Set(ctx, "key2", "value2", TTLShort)
	cache.Set(ctx, "key3", "value3", TTLShort)

	// Clear cache
	err := cache.Clear(ctx)
	assert.NoError(t, err)

	// Verify all keys are gone
	var result string
	err = cache.Get(ctx, "key1", &result)
	assert.Error(t, err)

	err = cache.Get(ctx, "key2", &result)
	assert.Error(t, err)

	err = cache.Get(ctx, "key3", &result)
	assert.Error(t, err)
}

func TestMemoryCache_GetStats(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// Set some keys
	cache.Set(ctx, "key1", "value1", TTLShort)
	cache.Set(ctx, "key2", "value2", TTLShort)

	stats := cache.GetStats()

	assert.NotNil(t, stats)
	assert.Equal(t, 2, stats["total_keys"])
	assert.GreaterOrEqual(t, stats["active_keys"], 0)
}

func TestCacheKeyBuilder(t *testing.T) {
	builder := NewCacheKeyBuilder("user")
	key := builder.Add("123").Add("profile").Build()

	assert.Equal(t, "user:123:profile", key)
}

func TestCacheKey(t *testing.T) {
	key := CacheKey("user", "123", "posts")
	assert.Equal(t, "user:123:posts", key)
}

func TestMatchesPattern(t *testing.T) {
	assert.True(t, matchesPattern("user:1", "user:*"))
	assert.True(t, matchesPattern("user:123", "user:*"))
	assert.False(t, matchesPattern("post:1", "user:*"))
	assert.True(t, matchesPattern("user:1", "user:1"))
	assert.False(t, matchesPattern("user:2", "user:1"))
}
