package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gofiber-template/domain/dto"
	"gofiber-template/domain/repositories"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// FeedCacheService handles caching for post feeds
type FeedCacheService struct {
	client *redis.Client
}

func NewFeedCacheService(redisClient *RedisClient) *FeedCacheService {
	return &FeedCacheService{
		client: redisClient.client,
	}
}

// Cache TTL configurations
const (
	HotFeedTTL  = 5 * time.Minute  // Hot posts change slowly
	NewFeedTTL  = 1 * time.Minute  // New posts change frequently
	TopFeedTTL  = 10 * time.Minute // Top posts change very slowly
	UserFeedTTL = 3 * time.Minute  // User's feed
	TagFeedTTL  = 5 * time.Minute  // Tag feeds
)

// ========== Feed Caching ==========

// CacheFeed caches a feed result with appropriate TTL
func (s *FeedCacheService) CacheFeed(ctx context.Context, cacheKey string, posts []dto.PostResponse, ttl time.Duration) error {
	data, err := json.Marshal(posts)
	if err != nil {
		return fmt.Errorf("failed to marshal posts: %w", err)
	}

	err = s.client.Set(ctx, cacheKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache feed: %w", err)
	}

	// Track cache write
	s.incrementCacheStat(ctx, "feed:cache:writes")

	return nil
}

// GetCachedFeed retrieves a cached feed
func (s *FeedCacheService) GetCachedFeed(ctx context.Context, cacheKey string) ([]dto.PostResponse, error) {
	data, err := s.client.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			// Cache miss - track it
			s.incrementCacheStat(ctx, "feed:cache:misses")
			return nil, nil // Cache miss, not an error
		}
		return nil, fmt.Errorf("failed to get cached feed: %w", err)
	}

	var posts []dto.PostResponse
	err = json.Unmarshal(data, &posts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached feed: %w", err)
	}

	// Cache hit - track it
	s.incrementCacheStat(ctx, "feed:cache:hits")

	return posts, nil
}

// InvalidateFeed invalidates a specific feed cache
func (s *FeedCacheService) InvalidateFeed(ctx context.Context, cacheKey string) error {
	return s.client.Del(ctx, cacheKey).Err()
}

// InvalidateAllFeeds invalidates all feed caches (use when post is created/deleted)
func (s *FeedCacheService) InvalidateAllFeeds(ctx context.Context) error {
	pattern := "feed:*"
	keys, err := s.scanKeys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}

	return nil
}

// InvalidateUserFeeds invalidates all feeds for a specific user
func (s *FeedCacheService) InvalidateUserFeeds(ctx context.Context, userID uuid.UUID) error {
	pattern := fmt.Sprintf("feed:user:%s:*", userID.String())
	keys, err := s.scanKeys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}

	return nil
}

// InvalidateTagFeeds invalidates all feeds for a specific tag
func (s *FeedCacheService) InvalidateTagFeeds(ctx context.Context, tagName string) error {
	pattern := fmt.Sprintf("feed:tag:%s:*", tagName)
	keys, err := s.scanKeys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}

	return nil
}

// ========== Cache Key Builders ==========

// BuildFeedCacheKey builds a cache key for main feed
func (s *FeedCacheService) BuildFeedCacheKey(sortBy repositories.PostSortBy, page, limit int) string {
	return fmt.Sprintf("feed:main:%s:page:%d:limit:%d", sortBy, page, limit)
}

// BuildUserFeedCacheKey builds a cache key for user's feed
func (s *FeedCacheService) BuildUserFeedCacheKey(userID uuid.UUID, page, limit int) string {
	return fmt.Sprintf("feed:user:%s:page:%d:limit:%d", userID.String(), page, limit)
}

// BuildTagFeedCacheKey builds a cache key for tag feed
func (s *FeedCacheService) BuildTagFeedCacheKey(tagName string, sortBy repositories.PostSortBy, page, limit int) string {
	return fmt.Sprintf("feed:tag:%s:%s:page:%d:limit:%d", tagName, sortBy, page, limit)
}

// GetFeedTTL returns appropriate TTL for a feed type
func (s *FeedCacheService) GetFeedTTL(sortBy repositories.PostSortBy) time.Duration {
	switch sortBy {
	case repositories.SortByHot:
		return HotFeedTTL
	case repositories.SortByNew:
		return NewFeedTTL
	case repositories.SortByTop:
		return TopFeedTTL
	default:
		return NewFeedTTL
	}
}

// ========== Cache Statistics ==========

// incrementCacheStat increments a cache statistic counter
func (s *FeedCacheService) incrementCacheStat(ctx context.Context, stat string) {
	// Non-blocking increment (ignore errors)
	s.client.Incr(ctx, stat)
}

// GetCacheStats retrieves cache statistics
func (s *FeedCacheService) GetCacheStats(ctx context.Context) (map[string]int64, error) {
	stats := map[string]int64{
		"hits":   0,
		"misses": 0,
		"writes": 0,
	}

	hits, _ := s.client.Get(ctx, "feed:cache:hits").Int64()
	misses, _ := s.client.Get(ctx, "feed:cache:misses").Int64()
	writes, _ := s.client.Get(ctx, "feed:cache:writes").Int64()

	stats["hits"] = hits
	stats["misses"] = misses
	stats["writes"] = writes

	// Calculate hit rate
	total := hits + misses
	if total > 0 {
		stats["hit_rate_percent"] = (hits * 100) / total
	}

	return stats, nil
}

// ResetCacheStats resets all cache statistics
func (s *FeedCacheService) ResetCacheStats(ctx context.Context) error {
	keys := []string{
		"feed:cache:hits",
		"feed:cache:misses",
		"feed:cache:writes",
	}
	return s.client.Del(ctx, keys...).Err()
}

// ========== Internal Helpers ==========

// scanKeys scans Redis for keys matching a pattern
func (s *FeedCacheService) scanKeys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	return keys, iter.Err()
}
