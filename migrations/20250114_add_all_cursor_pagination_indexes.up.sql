-- Comprehensive Migration: Add cursor pagination indexes for all tables
-- Date: 2025-01-14
-- Description: Create indexes for cursor-based pagination across all entities
-- Phase: ALL (Posts, Comments, Notifications, Follows, Saved Posts)

-- ============================================================================
-- PHASE 1: POSTS INDEXES (Already created in separate migration)
-- ============================================================================
-- See: 20250114_add_cursor_pagination_indexes.up.sql

-- ============================================================================
-- PHASE 2: COMMENTS INDEXES
-- ============================================================================

-- Index 1: Comments by Post (with sorting)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_comments_by_post_cursor
ON comments(post_id, parent_id, is_deleted, created_at DESC, id DESC)
WHERE is_deleted = false AND parent_id IS NULL;

-- Index 2: Comments by Author
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_comments_by_author_cursor
ON comments(author_id, is_deleted, created_at DESC, id DESC)
WHERE is_deleted = false;

-- Index 3: Replies to Comment
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_comments_replies_cursor
ON comments(parent_id, is_deleted, created_at DESC, id DESC)
WHERE is_deleted = false AND parent_id IS NOT NULL;

-- Index 4: Comments for Top sorting (by votes)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_comments_by_votes_cursor
ON comments(post_id, is_deleted, votes DESC, created_at DESC, id DESC)
WHERE is_deleted = false;

-- ============================================================================
-- PHASE 2: NOTIFICATIONS INDEXES
-- ============================================================================

-- Index 1: Notifications by User (all)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_by_user_cursor
ON notifications(user_id, created_at DESC, id DESC);

-- Index 2: Unread Notifications by User
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_unread_cursor
ON notifications(user_id, is_read, created_at DESC, id DESC)
WHERE is_read = false;

-- ============================================================================
-- PHASE 3: FOLLOWS INDEXES (Social Features)
-- ============================================================================

-- Index 1: Followers list
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_follows_followers_cursor
ON follows(following_id, created_at DESC, follower_id DESC);

-- Index 2: Following list
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_follows_following_cursor
ON follows(follower_id, created_at DESC, following_id DESC);

-- ============================================================================
-- PHASE 3: SAVED POSTS INDEXES
-- ============================================================================

-- Index 1: Saved posts by user
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_saved_posts_by_user_cursor
ON saved_posts(user_id, created_at DESC, post_id DESC);

-- ============================================================================
-- Analysis and Verification
-- ============================================================================

-- After creating indexes, verify they are being used:
--
-- For Comments:
-- EXPLAIN ANALYZE
-- SELECT * FROM comments
-- WHERE post_id = 'some-uuid' AND parent_id IS NULL AND is_deleted = false
-- AND (created_at, id) < ('2025-01-14 10:00:00', 'cursor-id')
-- ORDER BY created_at DESC, id DESC
-- LIMIT 20;
--
-- For Notifications:
-- EXPLAIN ANALYZE
-- SELECT * FROM notifications
-- WHERE user_id = 'some-uuid'
-- AND (created_at, id) < ('2025-01-14 10:00:00', 'cursor-id')
-- ORDER BY created_at DESC, id DESC
-- LIMIT 20;
--
-- For Follows:
-- EXPLAIN ANALYZE
-- SELECT users.* FROM follows
-- JOIN users ON follows.follower_id = users.id
-- WHERE follows.following_id = 'some-uuid'
-- AND (follows.created_at, follows.follower_id) < ('2025-01-14 10:00:00', 'cursor-id')
-- ORDER BY follows.created_at DESC, follows.follower_id DESC
-- LIMIT 20;

-- Check all cursor indexes:
SELECT
  schemaname,
  tablename,
  indexname,
  pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE indexname LIKE '%_cursor%'
ORDER BY pg_relation_size(indexrelid) DESC;
