package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryCache is an in-memory cache implementation
type MemoryCache struct {
	store map[string]*cacheItem
	mu    sync.RWMutex
}

type cacheItem struct {
	value      string
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		store: make(map[string]*cacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a value from cache
func (c *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	c.mu.RLock()
	item, exists := c.store[key]
	c.mu.RUnlock()

	if !exists {
		return NewCacheError("get", key, fmt.Errorf("key not found"))
	}

	// Check if expired
	if time.Now().After(item.expiration) {
		c.Delete(ctx, key)
		return NewCacheError("get", key, fmt.Errorf("key expired"))
	}

	// Deserialize value
	if err := DeserializeValue(item.value, dest); err != nil {
		return NewCacheError("get", key, err)
	}

	return nil
}

// Set stores a value in cache
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Serialize value
	serialized, err := SerializeValue(value)
	if err != nil {
		return NewCacheError("set", key, err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = &cacheItem{
		value:      serialized,
		expiration: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a key from cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, key)
	return nil
}

// DeletePattern deletes all keys matching a pattern
func (c *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple pattern matching (supports * wildcard at end)
	for key := range c.store {
		if matchesPattern(key, pattern) {
			delete(c.store, key)
		}
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	item, exists := c.store[key]
	c.mu.RUnlock()

	if !exists {
		return false, nil
	}

	// Check if expired
	if time.Now().After(item.expiration) {
		c.Delete(ctx, key)
		return false, nil
	}

	return true, nil
}

// Clear removes all items from cache
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store = make(map[string]*cacheItem)
	return nil
}

// GetStats returns cache statistics
func (c *MemoryCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalKeys := len(c.store)
	expiredKeys := 0

	now := time.Now()
	for _, item := range c.store {
		if now.After(item.expiration) {
			expiredKeys++
		}
	}

	return map[string]interface{}{
		"total_keys":   totalKeys,
		"active_keys":  totalKeys - expiredKeys,
		"expired_keys": expiredKeys,
	}
}

// cleanupExpired removes expired items periodically
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()

		for key, item := range c.store {
			if now.After(item.expiration) {
				delete(c.store, key)
			}
		}

		c.mu.Unlock()
	}
}

// matchesPattern checks if a key matches a pattern
func matchesPattern(key, pattern string) bool {
	// Simple implementation: only supports * at the end
	if len(pattern) == 0 {
		return false
	}

	if pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}

	return key == pattern
}
