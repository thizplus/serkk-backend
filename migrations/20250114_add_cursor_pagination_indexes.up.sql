-- Migration: Add composite indexes for cursor-based pagination
-- Date: 2025-01-14
-- Description: Create optimized indexes for post feed pagination with cursor support
--
-- These indexes are designed to support cursor-based pagination for different sort orders:
-- - New: created_at DESC
-- - Top: votes DESC
-- - Hot: calculated score DESC
--
-- Using CONCURRENTLY to avoid blocking production writes
-- Note: CONCURRENTLY requires running outside a transaction in PostgreSQL

-- ============================================================================
-- Index 1: New Feed (created_at DESC)
-- ============================================================================
-- Supports: GET /posts?sort=new
-- Query pattern: WHERE is_deleted = false AND status = 'published'
--                AND created_at < cursor_created_at
--                ORDER BY created_at DESC, id DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_feed_new
ON posts(is_deleted, status, created_at DESC, id DESC)
WHERE is_deleted = false AND status = 'published';

-- ============================================================================
-- Index 2: Top Feed (votes DESC)
-- ============================================================================
-- Supports: GET /posts?sort=top
-- Query pattern: WHERE is_deleted = false AND status = 'published'
--                AND votes <= cursor_votes
--                ORDER BY votes DESC, created_at DESC, id DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_feed_top
ON posts(is_deleted, status, votes DESC, created_at DESC, id DESC)
WHERE is_deleted = false AND status = 'published';

-- ============================================================================
-- Index 3: Hot Feed (recent posts only)
-- ============================================================================
-- Supports: GET /posts?sort=hot
-- Query pattern: WHERE is_deleted = false AND status = 'published'
--                AND created_at > NOW() - INTERVAL '7 days'
--                ORDER BY hot_score DESC, created_at DESC, id DESC
--
-- Note: This is a partial index for recent posts only (7 days)
-- Hot score is calculated at query time, this index helps with filtering and sorting
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_feed_hot
ON posts(is_deleted, status, created_at DESC, votes DESC, id DESC)
WHERE is_deleted = false
  AND status = 'published'
  AND created_at > NOW() - INTERVAL '7 days';

-- ============================================================================
-- Index 4: Posts by Author (for user profiles)
-- ============================================================================
-- Supports: GET /posts/author/:authorId
-- Query pattern: WHERE author_id = ? AND is_deleted = false AND status = 'published'
--                AND created_at < cursor_created_at
--                ORDER BY created_at DESC, id DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_by_author_cursor
ON posts(author_id, is_deleted, status, created_at DESC, id DESC)
WHERE is_deleted = false AND status = 'published';

-- ============================================================================
-- Index 5: Posts by Tag (requires post_tags join)
-- ============================================================================
-- Note: This index is on the posts table
-- The post_tags table already has indexes, but we need to optimize the posts side
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_for_tag_join
ON posts(is_deleted, status, created_at DESC, id DESC)
WHERE is_deleted = false AND status = 'published';

-- ============================================================================
-- Index 6: Posts for Feed (following users)
-- ============================================================================
-- Supports: GET /posts/feed (personalized feed for logged-in users)
-- This is used with a JOIN to the follows table
-- Query pattern: SELECT posts.* FROM posts
--                JOIN follows ON posts.author_id = follows.following_id
--                WHERE follows.follower_id = current_user_id
--                AND posts.created_at < cursor_created_at
--                ORDER BY posts.created_at DESC, posts.id DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_feed_following
ON posts(author_id, is_deleted, status, created_at DESC, id DESC)
WHERE is_deleted = false AND status = 'published';

-- ============================================================================
-- Analysis and Verification
-- ============================================================================
-- After creating indexes, verify they are being used:
--
-- EXPLAIN ANALYZE
-- SELECT * FROM posts
-- WHERE is_deleted = false AND status = 'published'
-- ORDER BY created_at DESC, id DESC
-- LIMIT 20;
--
-- Should show: Index Scan using idx_posts_feed_new
--
-- Check index sizes:
-- SELECT
--   schemaname,
--   tablename,
--   indexname,
--   pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
--   idx_scan as number_of_scans,
--   idx_tup_read as tuples_read,
--   idx_tup_fetch as tuples_fetched
-- FROM pg_stat_user_indexes
-- WHERE tablename = 'posts'
-- ORDER BY pg_relation_size(indexrelid) DESC;
