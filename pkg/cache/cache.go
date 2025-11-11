package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Cache defines the cache interface
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
}

// CacheKey generates a cache key with prefix
func CacheKey(prefix string, parts ...string) string {
	key := prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// CacheError represents a cache error
type CacheError struct {
	Operation string
	Key       string
	Err       error
}

func (e *CacheError) Error() string {
	return fmt.Sprintf("cache %s error for key '%s': %v", e.Operation, e.Key, e.Err)
}

// NewCacheError creates a new cache error
func NewCacheError(operation, key string, err error) *CacheError {
	return &CacheError{
		Operation: operation,
		Key:       key,
		Err:       err,
	}
}

// Common TTL constants
const (
	TTLShort  = 5 * time.Minute
	TTLMedium = 15 * time.Minute
	TTLLong   = 1 * time.Hour
	TTLDay    = 24 * time.Hour
)

// CacheKeyBuilder helps build cache keys
type CacheKeyBuilder struct {
	parts []string
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder(prefix string) *CacheKeyBuilder {
	return &CacheKeyBuilder{
		parts: []string{prefix},
	}
}

// Add adds a part to the cache key
func (b *CacheKeyBuilder) Add(part string) *CacheKeyBuilder {
	b.parts = append(b.parts, part)
	return b
}

// Build builds the final cache key
func (b *CacheKeyBuilder) Build() string {
	return CacheKey(b.parts[0], b.parts[1:]...)
}

// SerializeValue serializes a value to JSON for caching
func SerializeValue(value interface{}) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to serialize value: %w", err)
	}
	return string(data), nil
}

// DeserializeValue deserializes a JSON string to a value
func DeserializeValue(data string, dest interface{}) error {
	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to deserialize value: %w", err)
	}
	return nil
}
